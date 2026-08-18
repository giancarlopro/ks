package cmd

import (
	"os"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	clusterName := "test-cluster"
	registerCluster(t, clusterName)

	deleteCmd.Run(deleteCmd, []string{clusterName})

	if _, err := os.Stat(clusterConfigFile(clusterName)); !os.IsNotExist(err) {
		t.Fatalf("Configuration file not deleted: %s", clusterConfigFile(clusterName))
	}
}

func TestDeleteCommandUnknownCluster(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// A cluster that does not exist is reported, not created.
	deleteCmd.Run(deleteCmd, []string{"missing"})

	if _, err := os.Stat(clusterConfigFile("missing")); !os.IsNotExist(err) {
		t.Error("Expected no configuration file")
	}
}
