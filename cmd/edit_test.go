package cmd

import (
	"os"
	"testing"
)

func TestEditCluster(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	// `true` stands in for an editor. It exits at once and changes nothing.
	t.Setenv("EDITOR", "true")

	clusterName := "test-cluster"
	registerCluster(t, clusterName)

	before, err := os.ReadFile(clusterConfigFile(clusterName))
	if err != nil {
		t.Fatalf("Error reading config file: %v", err)
	}

	editCmd.Run(editCmd, []string{clusterName})

	after, err := os.ReadFile(clusterConfigFile(clusterName))
	if err != nil {
		t.Fatalf("Error reading config file: %v", err)
	}
	if string(after) != string(before) {
		t.Error("Expected an editor that changes nothing to leave the file alone")
	}
}
