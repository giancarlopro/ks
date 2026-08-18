package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/giancarlopro/ks/config"
	"gopkg.in/yaml.v2"
)

// sampleKubeconfig is a kubeconfig with one context. The context name does not
// match the registered name, so a test can prove the rename.
func sampleKubeconfig(name string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    server: https://%[1]s.example.com
  name: gke_project_region_%[1]s
contexts:
- context:
    cluster: gke_project_region_%[1]s
    namespace: default
    user: gke_project_region_%[1]s
  name: gke_project_region_%[1]s
current-context: gke_project_region_%[1]s
users:
- name: gke_project_region_%[1]s
  user:
    token: %[1]s-token
`, name)
}

// registerCluster writes a source kubeconfig for a cluster.
func registerCluster(t *testing.T, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(clusterConfigFile(name)), 0755); err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	if err := os.WriteFile(clusterConfigFile(name), []byte(sampleKubeconfig(name)), 0600); err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}
}

// currentContextOf reads current-context from a kubeconfig file.
func currentContextOf(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Error reading %s: %v", path, err)
	}
	var doc struct {
		CurrentContext string `yaml:"current-context"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Error parsing %s: %v", path, err)
	}
	return doc.CurrentContext
}

func TestActivateCluster(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Spawn a shell that exits at once, so the test does not wait for input.
	t.Setenv("SHELL", "/bin/true")

	registerCluster(t, "test-cluster")
	registerCluster(t, "other-cluster")

	if err := activateCluster("test-cluster"); err != nil {
		t.Fatalf("Error activating cluster: %v", err)
	}

	// Activation builds the merged kubeconfig and activates this cluster.
	merged := config.GeneratedConfigFile("test-cluster")
	if got := currentContextOf(t, merged); got != "test-cluster" {
		t.Errorf("Expected current-context test-cluster, got %q", got)
	}

	// Every registered cluster gets a merged kubeconfig of its own.
	if _, err := os.Stat(config.GeneratedConfigFile("other-cluster")); err != nil {
		t.Errorf("Expected a merged config for other-cluster: %v", err)
	}
}

func TestActivateClusterUnknown(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := activateCluster("missing"); err == nil {
		t.Error("Expected an error for a cluster that does not exist")
	}
}

func TestActivateClusterBroken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "/bin/true")

	registerCluster(t, "test-cluster")
	if err := os.WriteFile(clusterConfigFile("broken"), []byte("\tnot: [valid"), 0600); err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}

	// The broken cluster fails on its own.
	if err := activateCluster("broken"); err == nil {
		t.Error("Expected an error for a cluster that does not parse")
	}

	// It does not block the clusters that are fine.
	if err := activateCluster("test-cluster"); err != nil {
		t.Errorf("Expected the healthy cluster to activate: %v", err)
	}
}

func TestActivateClusterCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SHELL", "/bin/true")
	registerCluster(t, "test-cluster")

	// Run through the root command, so the test covers the cobra wiring. A
	// subprocess is not used: `go run` would put its module cache under the
	// temporary HOME, and the test could not clean that up.
	rootCmd.SetArgs([]string{"activate", "test-cluster"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Error running activate command: %v", err)
	}

	if got := currentContextOf(t, config.GeneratedConfigFile("test-cluster")); got != "test-cluster" {
		t.Errorf("Expected current-context test-cluster, got %q", got)
	}
}
