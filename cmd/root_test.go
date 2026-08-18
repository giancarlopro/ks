package cmd

import (
	"testing"

	"github.com/giancarlopro/ks/config"
)

func TestRootActivatesAClusterByName(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// Spawn a shell that exits at once, so the test does not wait for input.
	t.Setenv("SHELL", "/bin/true")

	registerCluster(t, "test-cluster")

	// `ks <cluster-name>` activates a cluster without the selector.
	rootCmd.SetArgs([]string{"test-cluster"})
	t.Cleanup(func() { rootCmd.SetArgs(nil) })

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Error activating cluster: %v", err)
	}

	if got := currentContextOf(t, config.GeneratedConfigFile("test-cluster")); got != "test-cluster" {
		t.Errorf("Expected current-context test-cluster, got %q", got)
	}
}
