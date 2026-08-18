package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// providerConfig is a kubeconfig as a user has it before ks. It holds three
// contexts, with the long names providers write.
const providerConfig = `apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: ZmFrZS1jYQ==
    server: https://prod.example.com
  name: gke_acme_us-east1_prod
- cluster:
    server: https://staging.example.com
  name: arn:aws:eks:us-east-1:123:cluster/staging
contexts:
- context:
    cluster: gke_acme_us-east1_prod
    namespace: payments
    user: gke_acme_us-east1_prod
  name: gke_acme_us-east1_prod
- context:
    cluster: arn:aws:eks:us-east-1:123:cluster/staging
    user: arn:aws:eks:us-east-1:123:cluster/staging
  name: arn:aws:eks:us-east-1:123:cluster/staging
- context:
    cluster: gke_acme_us-east1_prod
    namespace: kube-system
    user: gke_acme_us-east1_prod
  name: prod-admin
current-context: gke_acme_us-east1_prod
users:
- name: gke_acme_us-east1_prod
  user:
    token: prod-token
- name: arn:aws:eks:us-east-1:123:cluster/staging
  user:
    client-certificate-data: Y2VydA==
    client-key-data: a2V5
`

// writeFile writes a file and creates the directories above it.
func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// parseDoc renders a kubeconfig and parses it back for assertions.
func parseDoc(t *testing.T, doc yaml.MapSlice) mergedFile {
	t.Helper()
	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("rendering config: %v", err)
	}
	var parsed mergedFile
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parsing config: %v", err)
	}
	return parsed
}

func TestSuggestName(t *testing.T) {
	cases := map[string]string{
		"gke_acme_us-east1_prod":                    "prod",
		"arn:aws:eks:us-east-1:123:cluster/staging": "staging",
		"minikube":                    "minikube",
		"kind-local":                  "kind-local",
		"Prod Cluster":                "prod-cluster",
		"docker-desktop":              "docker-desktop",
		"cluster.local/admin@example": "admin-example",
	}

	for context, want := range cases {
		if got := SuggestName(context); got != want {
			t.Errorf("SuggestName(%q) = %q, want %q", context, got, want)
		}
	}
}

func TestReadSourcesSplitsEveryContext(t *testing.T) {
	home := useTempHome(t)
	path := filepath.Join(home, "kubeconfig")
	writeFile(t, path, providerConfig)

	contexts, warnings, err := ReadSources([]string{path})
	if err != nil {
		t.Fatalf("ReadSources: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
	if len(contexts) != 3 {
		t.Fatalf("expected 3 contexts, got %d", len(contexts))
	}

	// The suggested name is the short one a user wants to type.
	suggestions := []string{contexts[0].Suggested, contexts[1].Suggested, contexts[2].Suggested}
	want := "prod,staging,prod-admin"
	if got := strings.Join(suggestions, ","); got != want {
		t.Errorf("expected suggestions %q, got %q", want, got)
	}

	if contexts[0].Server != "https://prod.example.com" {
		t.Errorf("expected the server of the first context, got %q", contexts[0].Server)
	}

	// Each context becomes a kubeconfig that holds it alone, with the cluster
	// and the user it names.
	first := parseDoc(t, contexts[0].Config())
	if first.CurrentContext != "gke_acme_us-east1_prod" {
		t.Errorf("expected current-context to name the context, got %q", first.CurrentContext)
	}
	if len(first.Contexts) != 1 || len(first.Clusters) != 1 || len(first.Users) != 1 {
		t.Fatalf("expected one of each entry, got %d contexts, %d clusters, %d users",
			len(first.Contexts), len(first.Clusters), len(first.Users))
	}
	if _, _, namespace, _ := first.context("gke_acme_us-east1_prod"); namespace != "payments" {
		t.Errorf("expected namespace payments, got %q", namespace)
	}
	if body, _ := first.cluster("gke_acme_us-east1_prod"); body["server"] != "https://prod.example.com" {
		t.Errorf("expected the server to survive, got %v", body)
	}

	// A field the typed structs do not model survives the import.
	if body, _ := first.user("gke_acme_us-east1_prod"); body["token"] != "prod-token" {
		t.Errorf("expected the token to survive, got %v", body)
	}
	second := parseDoc(t, contexts[1].Config())
	if body, _ := second.user("arn:aws:eks:us-east-1:123:cluster/staging"); body["client-certificate-data"] != "Y2VydA==" {
		t.Errorf("expected the client certificate to survive, got %v", body)
	}
}

func TestReadSourcesCopiesSharedCluster(t *testing.T) {
	home := useTempHome(t)
	path := filepath.Join(home, "kubeconfig")
	writeFile(t, path, providerConfig)

	contexts, _, err := ReadSources([]string{path})
	if err != nil {
		t.Fatalf("ReadSources: %v", err)
	}

	// prod-admin shares a cluster and a user with the first context. Its file
	// must still hold both.
	admin := parseDoc(t, contexts[2].Config())
	if body, found := admin.cluster("gke_acme_us-east1_prod"); !found || body["server"] != "https://prod.example.com" {
		t.Errorf("expected the shared cluster to be copied, got %v", body)
	}
	if _, found := admin.user("gke_acme_us-east1_prod"); !found {
		t.Errorf("expected the shared user to be copied")
	}
	if _, _, namespace, _ := admin.context("prod-admin"); namespace != "kube-system" {
		t.Errorf("expected namespace kube-system, got %q", namespace)
	}
}

func TestReadSourcesWarnsAboutUnusableContext(t *testing.T) {
	home := useTempHome(t)
	path := filepath.Join(home, "kubeconfig")
	writeFile(t, path, strings.Replace(providerConfig, "    cluster: gke_acme_us-east1_prod\n    namespace: kube-system", "    cluster: gone\n    namespace: kube-system", 1))

	contexts, warnings, err := ReadSources([]string{path})
	if err != nil {
		t.Fatalf("ReadSources: %v", err)
	}
	if len(contexts) != 2 {
		t.Errorf("expected 2 usable contexts, got %d", len(contexts))
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], `unknown cluster "gone"`) {
		t.Errorf("expected a warning about the unknown cluster, got %v", warnings)
	}
}

func TestReadSourcesFailsOnBrokenFile(t *testing.T) {
	home := useTempHome(t)
	good := filepath.Join(home, "good")
	bad := filepath.Join(home, "bad")
	writeFile(t, good, providerConfig)
	writeFile(t, bad, "\tnot: [valid")

	if _, _, err := ReadSources([]string{good, bad}); err == nil {
		t.Error("expected an error for the broken file")
	}
}

func TestReadSourcesAsOneKeepsEveryContext(t *testing.T) {
	home := useTempHome(t)
	first := filepath.Join(home, "first")
	second := filepath.Join(home, "second")
	writeFile(t, first, providerConfig)
	writeFile(t, second, `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://lab.example.com
  name: lab
contexts:
- context:
    cluster: lab
    user: lab
  name: lab
current-context: lab
users:
- name: lab
  user:
    token: lab-token
`)

	doc, err := ReadSourcesAsOne([]string{first, second})
	if err != nil {
		t.Fatalf("ReadSourcesAsOne: %v", err)
	}

	parsed := parseDoc(t, doc)
	if len(parsed.Contexts) != 4 {
		t.Errorf("expected 4 contexts, got %d", len(parsed.Contexts))
	}
	// The first file that names a current-context wins.
	if parsed.CurrentContext != "gke_acme_us-east1_prod" {
		t.Errorf("expected current-context from the first file, got %q", parsed.CurrentContext)
	}
	if _, _, _, found := parsed.context("lab"); !found {
		t.Errorf("expected the context of the second file")
	}
}

func TestSourcePathsReadsKubeconfigVariable(t *testing.T) {
	home := useTempHome(t)
	first := filepath.Join(home, "first")
	second := filepath.Join(home, "second")
	writeFile(t, first, providerConfig)
	writeFile(t, second, providerConfig)

	t.Setenv("KUBECONFIG", strings.Join([]string{first, second}, string(filepath.ListSeparator)))

	paths, err := SourcePaths(nil)
	if err != nil {
		t.Fatalf("SourcePaths: %v", err)
	}
	if len(paths) != 2 || paths[0] != first || paths[1] != second {
		t.Errorf("expected both paths of KUBECONFIG, got %v", paths)
	}

	// An argument replaces the variable.
	paths, err = SourcePaths([]string{second})
	if err != nil {
		t.Fatalf("SourcePaths: %v", err)
	}
	if len(paths) != 1 || paths[0] != second {
		t.Errorf("expected the argument to win, got %v", paths)
	}
}

func TestSourcePathsFallsBackToDefaultConfig(t *testing.T) {
	useTempHome(t)
	t.Setenv("KUBECONFIG", "")
	writeFile(t, DefaultConfigFile(), providerConfig)

	paths, err := SourcePaths(nil)
	if err != nil {
		t.Fatalf("SourcePaths: %v", err)
	}
	if len(paths) != 1 || paths[0] != DefaultConfigFile() {
		t.Errorf("expected the default config, got %v", paths)
	}
}

func TestSourcePathsRefusesManagedFile(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	rebuild(t)

	// A generated file belongs to ks.
	if _, err := SourcePaths([]string{GeneratedConfigFile("prod")}); err == nil {
		t.Error("expected a generated file to be refused")
	}
	// So does a source file.
	if _, err := SourcePaths([]string{ClusterConfigFile("prod")}); err == nil {
		t.Error("expected a cluster file to be refused")
	}

	// A link into the ks directory belongs to ks as well.
	link := filepath.Join(HomeDir(), "link.yaml")
	if err := os.Symlink(GeneratedConfigFile("prod"), link); err != nil {
		t.Fatalf("creating link: %v", err)
	}
	if _, err := SourcePaths([]string{link}); err == nil {
		t.Error("expected a link into the ks directory to be refused")
	}
}

func TestSourcePathsReportsMissingFile(t *testing.T) {
	home := useTempHome(t)

	if _, err := SourcePaths([]string{filepath.Join(home, "gone")}); err == nil {
		t.Error("expected an error for a file that does not exist")
	}
}

func TestFreeNameAvoidsRegisteredNames(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))

	if got := FreeName("prod", nil); got != "prod-2" {
		t.Errorf("expected prod-2, got %q", got)
	}
	if got := FreeName("prod", map[string]bool{"prod-2": true}); got != "prod-3" {
		t.Errorf("expected prod-3, got %q", got)
	}
	if got := FreeName("dev", nil); got != "dev" {
		t.Errorf("expected dev, got %q", got)
	}
}
