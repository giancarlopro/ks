package ks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestActivateCluster(t *testing.T) {
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

	// Set the KUBECONFIG environment variable
	os.Setenv("KUBECONFIG", configFile)

	// Run the activateCluster function
	err = activateCluster(clusterName)
	if err != nil {
		t.Errorf("Error activating cluster: %v", err)
	}

	// Verify that the KUBECONFIG environment variable is set correctly
	expectedKubeconfig := configFile
	actualKubeconfig := os.Getenv("KUBECONFIG")
	if actualKubeconfig != expectedKubeconfig {
		t.Errorf("KUBECONFIG environment variable not set correctly. Expected: %s, Got: %s", expectedKubeconfig, actualKubeconfig)
	}
}

func TestActivateClusterCommand(t *testing.T) {
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

	// Set the KUBECONFIG environment variable
	os.Setenv("KUBECONFIG", configFile)

	// Run the activate command
	cmd := exec.Command("go", "run", ".", "activate", clusterName)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		t.Errorf("Error running activate command: %v", err)
	}

	// Verify that the KUBECONFIG environment variable is set correctly
	expectedKubeconfig := configFile
	actualKubeconfig := os.Getenv("KUBECONFIG")
	if actualKubeconfig != expectedKubeconfig {
		t.Errorf("KUBECONFIG environment variable not set correctly. Expected: %s, Got: %s", expectedKubeconfig, actualKubeconfig)
	}
}
