package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetDefaultCluster(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "clusters")
	os.MkdirAll(configDir, 0755)

	// Create a sample cluster configuration file
	clusterName := "test-cluster"
	configFile := filepath.Join(configDir, clusterName+".yaml")
	os.WriteFile(configFile, []byte("sample config"), 0644)

	// Create the .kube directory if it does not exist
	kubeDir := filepath.Join(tempDir, ".kube")
	os.MkdirAll(kubeDir, 0755)

	// Set the environment variable for the config directory
	os.Setenv("HOME", tempDir)

	// Call the setDefaultCluster function
	err := setDefaultCluster(clusterName)
	if err != nil {
		t.Fatalf("Error setting default cluster: %v", err)
	}

	// Check if the symbolic link was created correctly
	defaultConfigFile := filepath.Join(tempDir, ".kube", "config")
	linkTarget, err := os.Readlink(defaultConfigFile)
	if err != nil {
		t.Fatalf("Error reading symbolic link: %v", err)
	}

	if linkTarget != configFile {
		t.Errorf("Expected symbolic link target to be %s, but got %s", configFile, linkTarget)
	}
}
