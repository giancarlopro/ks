package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZshIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a .ksconfig file in the temporary directory
	ksconfigPath := filepath.Join(tempDir, ".ksconfig")
	clusterName := "test-cluster"
	err := os.WriteFile(ksconfigPath, []byte(clusterName), 0644)
	if err != nil {
		t.Fatalf("Error creating .ksconfig file: %v", err)
	}

	// Create a cluster configuration file in the temporary directory
	configDir := filepath.Join(tempDir, ".config", "ks", "clusters")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	configFile := filepath.Join(configDir, clusterName+".yaml")
	err = os.WriteFile(configFile, []byte("test-config"), 0644)
	if err != nil {
		t.Fatalf("Error creating cluster configuration file: %v", err)
	}

	// Set the current working directory to the temporary directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Error changing directory: %v", err)
	}

	// Run the zsh-integration command
	err = zshIntegrationCmd.RunE(zshIntegrationCmd, []string{})
	if err != nil {
		t.Fatalf("Error running zsh-integration command: %v", err)
	}

	// Check if the KUBECONFIG environment variable is set correctly
	expectedKubeconfig := configFile
	actualKubeconfig := os.Getenv("KUBECONFIG")
	if actualKubeconfig != expectedKubeconfig {
		t.Errorf("KUBECONFIG environment variable not set correctly. Expected: %s, Got: %s", expectedKubeconfig, actualKubeconfig)
	}
}
