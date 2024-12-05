package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddCommand(t *testing.T) {
	// Set up test environment
	homeDir := os.Getenv("HOME")
	configDir := filepath.Join(homeDir, ".config", "ks", "clusters")
	clusterName := "test-cluster"
	configFile := filepath.Join(configDir, clusterName+".yaml")

	// Create the directory if it does not exist
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}

	// Clean up any existing test files
	os.Remove(configFile)

	// Run the add command
	addCmd.SetArgs([]string{clusterName})
	err = addCmd.Execute()
	if err != nil {
		t.Fatalf("Error executing add command: %v", err)
	}

	// Check if the configuration file was created
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatalf("Configuration file not created: %s", configFile)
	}

	// Clean up test files
	os.Remove(configFile)
}
