package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestAddCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// The add command opens an editor. `true` exits at once and changes
	// nothing, so the template is what gets saved.
	t.Setenv("EDITOR", "true")

	clusterName := "test-cluster"
	addCmd.Run(addCmd, []string{clusterName})

	content, err := os.ReadFile(clusterConfigFile(clusterName))
	if err != nil {
		t.Fatalf("Configuration file not created: %v", err)
	}
	if !strings.Contains(string(content), "name: "+clusterName) {
		t.Errorf("Expected the template to name the cluster, got:\n%s", content)
	}
}
