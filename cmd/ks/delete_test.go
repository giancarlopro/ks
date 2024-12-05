package ks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	// Set up test environment
	homeDir := os.Getenv("HOME")
	configDir := filepath.Join(homeDir, ".config", "ks", "clusters")
	clusterName := "test-cluster"
	configFile := filepath.Join(configDir, clusterName+".yaml")

	// Create a temporary configuration file
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	defer os.RemoveAll(configDir)

	file, err := os.Create(configFile)
	if err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}
	defer file.Close()

	// Run the delete command
	deleteCmd.SetArgs([]string{clusterName})
	err = deleteCmd.Execute()
	if err != nil {
		t.Fatalf("Error executing delete command: %v", err)
	}

	// Check if the configuration file was deleted
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Fatalf("Configuration file not deleted: %s", configFile)
	}
}
