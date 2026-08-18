package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

// mergedFile is the merged kubeconfig, parsed back for assertions.
type mergedFile struct {
	APIVersion     string `yaml:"apiVersion"`
	Kind           string `yaml:"kind"`
	CurrentContext string `yaml:"current-context"`
	Clusters       []struct {
		Name    string                 `yaml:"name"`
		Cluster map[string]interface{} `yaml:"cluster"`
	} `yaml:"clusters"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster   string `yaml:"cluster"`
			User      string `yaml:"user"`
			Namespace string `yaml:"namespace"`
		} `yaml:"context"`
	} `yaml:"contexts"`
	Users []struct {
		Name string                 `yaml:"name"`
		User map[string]interface{} `yaml:"user"`
	} `yaml:"users"`
}

func (m mergedFile) contextNames() []string {
	names := make([]string, 0, len(m.Contexts))
	for _, ctx := range m.Contexts {
		names = append(names, ctx.Name)
	}
	return names
}

func (m mergedFile) context(name string) (string, string, string, bool) {
	for _, ctx := range m.Contexts {
		if ctx.Name == name {
			return ctx.Context.Cluster, ctx.Context.User, ctx.Context.Namespace, true
		}
	}
	return "", "", "", false
}

func (m mergedFile) cluster(name string) (map[string]interface{}, bool) {
	for _, c := range m.Clusters {
		if c.Name == name {
			return c.Cluster, true
		}
	}
	return nil, false
}

func (m mergedFile) user(name string) (map[string]interface{}, bool) {
	for _, u := range m.Users {
		if u.Name == name {
			return u.User, true
		}
	}
	return nil, false
}

// useTempHome points HOME at a temporary directory, so every ks path lives
// inside the test.
func useTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

// writeSource registers a cluster with the given kubeconfig body.
func writeSource(t *testing.T, name, body string) {
	t.Helper()
	if err := os.MkdirAll(ClustersDir(), 0755); err != nil {
		t.Fatalf("creating clusters dir: %v", err)
	}
	if err := os.WriteFile(ClusterConfigFile(name), []byte(body), 0600); err != nil {
		t.Fatalf("writing source %q: %v", name, err)
	}
}

// readMerged parses the merged kubeconfig of a cluster.
func readMerged(t *testing.T, name string) mergedFile {
	t.Helper()
	data, err := os.ReadFile(GeneratedConfigFile(name))
	if err != nil {
		t.Fatalf("reading merged config %q: %v", name, err)
	}
	var merged mergedFile
	if err := yaml.Unmarshal(data, &merged); err != nil {
		t.Fatalf("parsing merged config %q: %v", name, err)
	}
	return merged
}

// rebuild runs a rebuild and returns the warnings it printed.
func rebuild(t *testing.T) (*RebuildResult, string) {
	t.Helper()
	var warnings bytes.Buffer
	result, err := Rebuild(&warnings)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	return result, warnings.String()
}

// oneContext is a kubeconfig with a single context, as a cloud provider writes
// it. The names are long and do not match the registered name.
func oneContext(prefix string) string {
	return `apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: ` + prefix + `-ca
    server: https://` + prefix + `.example.com
  name: gke_project_region_` + prefix + `
contexts:
- context:
    cluster: gke_project_region_` + prefix + `
    namespace: default
    user: gke_project_region_` + prefix + `
  name: gke_project_region_` + prefix + `
current-context: gke_project_region_` + prefix + `
users:
- name: gke_project_region_` + prefix + `
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: gke-gcloud-auth-plugin
      provideClusterInfo: true
`
}

func TestRebuildRenamesSingleContext(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))

	result, warnings := rebuild(t)

	if warnings != "" {
		t.Errorf("expected no warnings, got %q", warnings)
	}
	if _, err := result.Path("prod"); err != nil {
		t.Fatalf("expected prod to build: %v", err)
	}

	merged := readMerged(t, "prod")

	if merged.APIVersion != "v1" || merged.Kind != "Config" {
		t.Errorf("expected apiVersion v1 and kind Config, got %q and %q", merged.APIVersion, merged.Kind)
	}
	if merged.CurrentContext != "prod" {
		t.Errorf("expected current-context prod, got %q", merged.CurrentContext)
	}

	cluster, user, namespace, found := merged.context("prod")
	if !found {
		t.Fatalf("expected a context named prod, got %v", merged.contextNames())
	}
	if cluster != "prod" || user != "prod" {
		t.Errorf("expected context prod to name cluster and user prod, got %q and %q", cluster, user)
	}
	if namespace != "default" {
		t.Errorf("expected namespace default, got %q", namespace)
	}

	body, found := merged.cluster("prod")
	if !found {
		t.Fatalf("expected a cluster named prod")
	}
	if body["server"] != "https://prod.example.com" {
		t.Errorf("expected the server to survive, got %v", body["server"])
	}
	if _, found := merged.user("prod"); !found {
		t.Errorf("expected a user named prod")
	}
}

func TestRebuildHoldsEveryRegisteredCluster(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	writeSource(t, "dev", oneContext("dev"))

	rebuild(t)

	for _, name := range []string{"prod", "dev"} {
		merged := readMerged(t, name)
		if merged.CurrentContext != name {
			t.Errorf("expected current-context %q, got %q", name, merged.CurrentContext)
		}
		if got := strings.Join(merged.contextNames(), ","); got != "dev,prod" {
			t.Errorf("expected contexts dev,prod in %q, got %q", name, got)
		}
		if len(merged.Clusters) != 2 || len(merged.Users) != 2 {
			t.Errorf("expected 2 clusters and 2 users in %q, got %d and %d", name, len(merged.Clusters), len(merged.Users))
		}
	}
}

// threeContexts is a kubeconfig that holds several contexts, as a hand-written
// file does.
const threeContexts = `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://one.example.com
  name: one
- cluster:
    server: https://two.example.com
  name: two
contexts:
- context:
    cluster: one
    user: one
  name: alpha
- context:
    cluster: two
    user: two
  name: beta
- context:
    cluster: one
    namespace: kube-system
    user: one
  name: gamma
current-context: beta
users:
- name: one
  user:
    token: one-token
- name: two
  user:
    token: two-token
`

func TestRebuildNamesExtraContexts(t *testing.T) {
	useTempHome(t)
	writeSource(t, "lab", threeContexts)

	rebuild(t)
	merged := readMerged(t, "lab")

	want := "lab,lab/alpha,lab/gamma"
	if got := strings.Join(merged.contextNames(), ","); got != want {
		t.Errorf("expected contexts %q, got %q", want, got)
	}
	if merged.CurrentContext != "lab" {
		t.Errorf("expected current-context lab, got %q", merged.CurrentContext)
	}

	// The bare name comes from current-context, which points at cluster two.
	cluster, user, _, _ := merged.context("lab")
	if cluster != "lab" || user != "lab" {
		t.Errorf("expected context lab to name cluster and user lab, got %q and %q", cluster, user)
	}
	body, found := merged.cluster("lab")
	if !found || body["server"] != "https://two.example.com" {
		t.Errorf("expected cluster lab to keep the server of context beta, got %v", body)
	}

	// alpha and gamma share one source cluster, so each gets its own entry.
	alpha, foundAlpha := merged.cluster("lab/alpha")
	gamma, foundGamma := merged.cluster("lab/gamma")
	if !foundAlpha || !foundGamma {
		t.Fatalf("expected an entry for each extra context")
	}
	if alpha["server"] != "https://one.example.com" || gamma["server"] != "https://one.example.com" {
		t.Errorf("expected both entries to keep the shared server, got %v and %v", alpha, gamma)
	}

	if _, _, namespace, _ := merged.context("lab/gamma"); namespace != "kube-system" {
		t.Errorf("expected namespace kube-system, got %q", namespace)
	}
}

func TestRebuildWithoutCurrentContextNamesNoBareContext(t *testing.T) {
	useTempHome(t)
	writeSource(t, "lab", strings.Replace(threeContexts, "current-context: beta", "current-context: gone", 1))

	rebuild(t)
	merged := readMerged(t, "lab")

	want := "lab/alpha,lab/beta,lab/gamma"
	if got := strings.Join(merged.contextNames(), ","); got != want {
		t.Errorf("expected contexts %q, got %q", want, got)
	}
	if merged.CurrentContext != "lab/alpha" {
		t.Errorf("expected current-context lab/alpha, got %q", merged.CurrentContext)
	}
}

func TestRebuildKeepsUnknownFields(t *testing.T) {
	useTempHome(t)
	writeSource(t, "lab", threeContexts)

	rebuild(t)
	merged := readMerged(t, "lab")

	// The typed structs in config.go only model an exec user. A token must
	// still survive the merge.
	user, found := merged.user("lab")
	if !found {
		t.Fatalf("expected a user named lab")
	}
	if user["token"] != "two-token" {
		t.Errorf("expected the token to survive, got %v", user)
	}
}

func TestRebuildSkipsBrokenSource(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	writeSource(t, "broken", "\tnot: [valid")

	result, warnings := rebuild(t)

	if !strings.Contains(warnings, "ks: skipping") || !strings.Contains(warnings, "broken.yaml") {
		t.Errorf("expected a warning about broken.yaml, got %q", warnings)
	}
	if _, err := result.Path("broken"); err == nil {
		t.Errorf("expected broken to report an error")
	}
	if _, err := result.Path("prod"); err != nil {
		t.Errorf("expected prod to build anyway: %v", err)
	}

	merged := readMerged(t, "prod")
	if got := strings.Join(merged.contextNames(), ","); got != "prod" {
		t.Errorf("expected only the prod context, got %q", got)
	}
}

func TestRebuildSkipsSourceWithoutContext(t *testing.T) {
	useTempHome(t)
	writeSource(t, "empty", "apiVersion: v1\nkind: Config\n")

	result, warnings := rebuild(t)

	if !strings.Contains(warnings, "no usable context") {
		t.Errorf("expected a warning about the missing context, got %q", warnings)
	}
	if _, err := result.Path("empty"); err == nil {
		t.Errorf("expected empty to report an error")
	}
	if _, err := os.Stat(GeneratedConfigFile("empty")); !os.IsNotExist(err) {
		t.Errorf("expected no merged config for empty")
	}
}

func TestRebuildSkipsContextWithUnknownReference(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	writeSource(t, "lab", strings.Replace(threeContexts, "    cluster: one\n    user: one\n  name: alpha", "    cluster: missing\n    user: one\n  name: alpha", 1))

	_, warnings := rebuild(t)

	if !strings.Contains(warnings, `context "alpha" names unknown cluster "missing"`) {
		t.Errorf("expected a warning about the unknown cluster, got %q", warnings)
	}

	merged := readMerged(t, "lab")
	want := "lab,lab/gamma,prod"
	if got := strings.Join(merged.contextNames(), ","); got != want {
		t.Errorf("expected contexts %q, got %q", want, got)
	}
}

func TestRebuildCarriesNamespaceForward(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	rebuild(t)

	// kubectl config set-context --current --namespace writes into the merged
	// file. A rebuild must keep that namespace.
	setNamespace(t, GeneratedConfigFile("prod"), "prod", "payments")

	writeSource(t, "dev", oneContext("dev"))
	rebuild(t)

	merged := readMerged(t, "prod")
	if _, _, namespace, _ := merged.context("prod"); namespace != "payments" {
		t.Errorf("expected namespace payments, got %q", namespace)
	}

	// The namespace belongs to the file the user changed, not to every file.
	other := readMerged(t, "dev")
	if _, _, namespace, _ := other.context("prod"); namespace != "default" {
		t.Errorf("expected the dev file to keep namespace default, got %q", namespace)
	}
}

func TestRebuildPrunesMergedConfigOfDeletedCluster(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	writeSource(t, "dev", oneContext("dev"))
	rebuild(t)

	if err := os.Remove(ClusterConfigFile("dev")); err != nil {
		t.Fatalf("removing source: %v", err)
	}
	rebuild(t)

	if _, err := os.Stat(GeneratedConfigFile("dev")); !os.IsNotExist(err) {
		t.Errorf("expected the merged config of dev to be removed")
	}
	merged := readMerged(t, "prod")
	if got := strings.Join(merged.contextNames(), ","); got != "prod" {
		t.Errorf("expected only the prod context, got %q", got)
	}
}

func TestRebuildLeavesNoTemporaryFile(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	rebuild(t)

	entries, err := os.ReadDir(GeneratedDir())
	if err != nil {
		t.Fatalf("reading generated dir: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() != "prod.yaml" {
			t.Errorf("unexpected file %q in the generated directory", entry.Name())
		}
	}

	info, err := os.Stat(GeneratedConfigFile("prod"))
	if err != nil {
		t.Fatalf("reading merged config: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %v", info.Mode().Perm())
	}
}

func TestRebuildIgnoresNonYAMLEntries(t *testing.T) {
	useTempHome(t)
	writeSource(t, "prod", oneContext("prod"))
	if err := os.WriteFile(filepath.Join(ClustersDir(), "notes.txt"), []byte("hello"), 0600); err != nil {
		t.Fatalf("writing note: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(ClustersDir(), "backups"), 0755); err != nil {
		t.Fatalf("creating backups dir: %v", err)
	}

	result, warnings := rebuild(t)

	if warnings != "" {
		t.Errorf("expected no warnings, got %q", warnings)
	}
	if len(result.Built) != 1 {
		t.Errorf("expected one merged config, got %d", len(result.Built))
	}
}

// setNamespace changes the namespace of one context, the way
// `kubectl config set-context --current --namespace` does.
func setNamespace(t *testing.T, path, context, namespace string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	var doc yaml.MapSlice
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	for _, item := range msList(doc, "contexts") {
		entry := item.(yaml.MapSlice)
		if msString(entry, "name") != context {
			continue
		}
		body, _ := msGet(entry, "context")
		msSet(entry, "context", msSet(body.(yaml.MapSlice), "namespace", namespace))
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("rendering %s: %v", path, err)
	}
	if err := os.WriteFile(path, out, 0600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}
