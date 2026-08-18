package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/giancarlopro/ks/config"
)

// providerConfig is a kubeconfig as a user has it before ks.
const providerConfig = `apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    server: https://prod.example.com
  name: gke_acme_us-east1_prod
- cluster:
    server: https://staging.example.com
  name: gke_acme_us-east1_staging
contexts:
- context:
    cluster: gke_acme_us-east1_prod
    namespace: payments
    user: gke_acme_us-east1_prod
  name: gke_acme_us-east1_prod
- context:
    cluster: gke_acme_us-east1_staging
    user: gke_acme_us-east1_staging
  name: gke_acme_us-east1_staging
current-context: gke_acme_us-east1_prod
users:
- name: gke_acme_us-east1_prod
  user:
    token: prod-token
- name: gke_acme_us-east1_staging
  user:
    token: staging-token
`

// importWithoutPrompts runs an import that accepts every suggestion, so the
// test needs no input. It resets the flags afterwards.
func importWithoutPrompts(t *testing.T, as string, args ...string) error {
	t.Helper()

	importYes, importAs = true, as
	t.Cleanup(func() { importYes, importAs = false, "" })

	return importKubeconfig(args)
}

// writeDefaultKubeconfig puts a kubeconfig at ~/.kube/config.
func writeDefaultKubeconfig(t *testing.T, body string) {
	t.Helper()
	if err := os.MkdirAll(config.KubeDir(), 0755); err != nil {
		t.Fatalf("creating kube dir: %v", err)
	}
	if err := os.WriteFile(config.DefaultConfigFile(), []byte(body), 0600); err != nil {
		t.Fatalf("writing default config: %v", err)
	}
}

// registeredNames lists the clusters ks knows about.
func registeredClusters(t *testing.T) []string {
	t.Helper()
	names, err := listClusters()
	if err != nil {
		t.Fatalf("listing clusters: %v", err)
	}
	sort.Strings(names)
	return names
}

func TestImportRegistersEveryContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	if err := importWithoutPrompts(t, ""); err != nil {
		t.Fatalf("import: %v", err)
	}

	// Each context becomes a cluster, under the short name ks suggests.
	if got := strings.Join(registeredClusters(t), ","); got != "prod,staging" {
		t.Errorf("expected clusters prod,staging, got %q", got)
	}

	// The kubeconfig moved aside, and the default now points at ks.
	if _, err := os.Stat(config.KubeBackupFile()); err != nil {
		t.Errorf("expected a backup at %s: %v", config.KubeBackupFile(), err)
	}
	target, err := os.Readlink(config.DefaultConfigFile())
	if err != nil {
		t.Fatalf("expected a link at %s: %v", config.DefaultConfigFile(), err)
	}
	// The active context was the prod one, so prod is the default.
	if target != config.GeneratedConfigFile("prod") {
		t.Errorf("expected the link to point at prod, got %s", target)
	}
	if got := currentContextOf(t, config.GeneratedConfigFile("prod")); got != "prod" {
		t.Errorf("expected current-context prod, got %q", got)
	}

	// The backup holds the original file.
	backup, err := os.ReadFile(config.KubeBackupFile())
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(backup) != providerConfig {
		t.Error("expected the backup to hold the original kubeconfig")
	}
}

// importWithAnswers runs an import that reads the given prompt answers.
func importWithAnswers(t *testing.T, answers string, args ...string) error {
	t.Helper()

	promptIn = strings.NewReader(answers)
	t.Cleanup(func() { promptIn = os.Stdin })

	return importKubeconfig(args)
}

func TestImportPromptsForEveryName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	// Enter accepts the suggestion, then a typed name replaces it, then the
	// move is confirmed.
	if err := importWithAnswers(t, "\nstg\ny\n"); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "prod,stg" {
		t.Errorf("expected clusters prod,stg, got %q", got)
	}
	if _, err := os.Readlink(config.DefaultConfigFile()); err != nil {
		t.Errorf("expected the default config to be linked: %v", err)
	}
}

func TestImportSkipsAContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	// A dash skips the first context. The active context is therefore not
	// registered, so the default config stays as it is.
	if err := importWithAnswers(t, "-\n\n"); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "staging" {
		t.Errorf("expected only staging, got %q", got)
	}
	if config.Managed(config.DefaultConfigFile()) {
		t.Error("expected the default config to be left alone")
	}
	if _, err := os.Stat(config.KubeBackupFile()); !os.IsNotExist(err) {
		t.Error("expected no backup")
	}
}

func TestImportSkipsEveryContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	// Skipping everything registers nothing, and is not an error.
	if err := importWithAnswers(t, "-\n-\n"); err != nil {
		t.Fatalf("import: %v", err)
	}

	if names := registeredClusters(t); len(names) != 0 {
		t.Errorf("expected no cluster, got %v", names)
	}
	if config.Managed(config.DefaultConfigFile()) {
		t.Error("expected the default config to be left alone")
	}
}

func TestImportReportsAKubeconfigWithoutContexts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, "apiVersion: v1\nkind: Config\n")

	if err := importWithAnswers(t, ""); err == nil {
		t.Error("expected an error for a kubeconfig with no context")
	}
}

func TestImportAsksAgainForATakenName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)
	registerCluster(t, "prod")

	// The suggestion is taken, so the prompt asks again. The second answer is
	// free.
	if err := importWithAnswers(t, "\nprod-live\n-\nn\n"); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "prod,prod-live" {
		t.Errorf("expected clusters prod,prod-live, got %q", got)
	}
}

func TestImportKeepsTheDefaultConfigOnNo(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	// A no answer to the move keeps the clusters and leaves the file alone.
	if err := importWithAnswers(t, "\n\nn\n"); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "prod,staging" {
		t.Errorf("expected clusters prod,staging, got %q", got)
	}
	if _, err := os.Stat(config.KubeBackupFile()); !os.IsNotExist(err) {
		t.Error("expected no backup after a no answer")
	}
	content, err := os.ReadFile(config.DefaultConfigFile())
	if err != nil {
		t.Fatalf("reading default config: %v", err)
	}
	if string(content) != providerConfig {
		t.Error("expected the default config to be untouched")
	}
}

func TestImportNumbersATakenName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)
	registerCluster(t, "prod")

	if err := importWithoutPrompts(t, ""); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "prod,prod-2,staging" {
		t.Errorf("expected the taken name to be numbered, got %q", got)
	}

	// The cluster that was already registered keeps its own content.
	if got := currentContextOf(t, clusterConfigFile("prod")); got != "gke_project_region_prod" {
		t.Errorf("expected the existing cluster to be untouched, got %q", got)
	}
}

func TestImportAsOneCluster(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	if err := importWithoutPrompts(t, "work"); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "work" {
		t.Errorf("expected one cluster named work, got %q", got)
	}

	// The merged config names the extra contexts under the base name.
	merged := config.GeneratedConfigFile("work")
	data, err := os.ReadFile(merged)
	if err != nil {
		t.Fatalf("reading merged config: %v", err)
	}
	for _, name := range []string{"name: work", "name: work/gke_acme_us-east1_staging"} {
		if !strings.Contains(string(data), name) {
			t.Errorf("expected %q in the merged config", name)
		}
	}
	if got := currentContextOf(t, merged); got != "work" {
		t.Errorf("expected current-context work, got %q", got)
	}
}

func TestImportReadsKubeconfigVariable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	other := filepath.Join(home, "work.yaml")
	if err := os.WriteFile(other, []byte(providerConfig), 0600); err != nil {
		t.Fatalf("writing source: %v", err)
	}
	t.Setenv("KUBECONFIG", other)

	if err := importWithoutPrompts(t, ""); err != nil {
		t.Fatalf("import: %v", err)
	}

	if got := strings.Join(registeredClusters(t), ","); got != "prod,staging" {
		t.Errorf("expected clusters prod,staging, got %q", got)
	}

	// A file that is not the default kubectl config is read, never moved.
	if _, err := os.Stat(other); err != nil {
		t.Errorf("expected the source file to stay: %v", err)
	}
	if _, err := os.Stat(config.KubeBackupFile()); !os.IsNotExist(err) {
		t.Errorf("expected no backup, because there was no default config")
	}
}

func TestImportRefusesManagedFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	registerCluster(t, "prod")

	if err := importWithoutPrompts(t, "", clusterConfigFile("prod")); err == nil {
		t.Error("expected a file that belongs to ks to be refused")
	}
}

func TestImportKeepsAnEarlierBackup(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)
	if err := os.WriteFile(config.KubeBackupFile(), []byte("earlier backup"), 0600); err != nil {
		t.Fatalf("writing backup: %v", err)
	}

	if err := importWithoutPrompts(t, ""); err != nil {
		t.Fatalf("import: %v", err)
	}

	// The clusters are registered, but the earlier backup is safe.
	if got := strings.Join(registeredClusters(t), ","); got != "prod,staging" {
		t.Errorf("expected clusters prod,staging, got %q", got)
	}
	backup, err := os.ReadFile(config.KubeBackupFile())
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(backup) != "earlier backup" {
		t.Error("expected the earlier backup to be kept")
	}
	// The default config is untouched, so it is still the user's own file.
	if config.Managed(config.DefaultConfigFile()) {
		t.Error("expected the default config to be left alone")
	}
}

func TestImportWritesNothingWhenASourceIsBroken(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	broken := filepath.Join(home, "broken.yaml")
	if err := os.WriteFile(broken, []byte("\tnot: [valid"), 0600); err != nil {
		t.Fatalf("writing source: %v", err)
	}

	if err := importWithoutPrompts(t, "", config.DefaultConfigFile(), broken); err == nil {
		t.Fatal("expected an error for the broken file")
	}

	if names := registeredClusters(t); len(names) != 0 {
		t.Errorf("expected no cluster to be registered, got %v", names)
	}
	if _, err := os.Stat(config.KubeBackupFile()); !os.IsNotExist(err) {
		t.Error("expected no backup after a failed import")
	}
}

func TestRestoreKubeconfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KUBECONFIG", "")
	writeDefaultKubeconfig(t, providerConfig)

	if err := importWithoutPrompts(t, ""); err != nil {
		t.Fatalf("import: %v", err)
	}
	if err := restoreKubeconfig(); err != nil {
		t.Fatalf("restore: %v", err)
	}

	// The original file is back, and the backup is gone.
	content, err := os.ReadFile(config.DefaultConfigFile())
	if err != nil {
		t.Fatalf("reading default config: %v", err)
	}
	if string(content) != providerConfig {
		t.Error("expected the original kubeconfig to be restored")
	}
	if _, err := os.Lstat(config.KubeBackupFile()); !os.IsNotExist(err) {
		t.Error("expected the backup to be consumed")
	}
	if _, err := os.Stat(config.DefaultRecordFile()); !os.IsNotExist(err) {
		t.Error("expected the default cluster record to be cleared")
	}

	// Nothing that import registered is lost.
	if got := strings.Join(registeredClusters(t), ","); got != "prod,staging" {
		t.Errorf("expected the imported clusters to stay, got %q", got)
	}
}

func TestRestoreKubeconfigWithoutBackup(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := restoreKubeconfig(); err == nil {
		t.Error("expected an error when there is no backup")
	}
}

func TestRestoreKubeconfigKeepsAFileKsDidNotCreate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	writeDefaultKubeconfig(t, providerConfig)
	if err := os.WriteFile(config.KubeBackupFile(), []byte("older config"), 0600); err != nil {
		t.Fatalf("writing backup: %v", err)
	}

	if err := restoreKubeconfig(); err == nil {
		t.Fatal("expected an error for a default config ks did not create")
	}

	restoreForce = true
	t.Cleanup(func() { restoreForce = false })

	if err := restoreKubeconfig(); err != nil {
		t.Fatalf("restore with force: %v", err)
	}
	content, err := os.ReadFile(config.DefaultConfigFile())
	if err != nil {
		t.Fatalf("reading default config: %v", err)
	}
	if string(content) != "older config" {
		t.Error("expected force to overwrite the default config")
	}
}
