package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEditCluster(t *testing.T) {
	// Set up test environment
	homeDir := os.Getenv("HOME")
	configDir := filepath.Join(homeDir, ".config", "ks", "clusters")
	clusterName := "test-cluster"
	configFile := filepath.Join(configDir, clusterName+".yaml")

	// Create a dummy config file
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	file, err := os.Create(configFile)
	if err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}
	file.Close()

	// Set the EDITOR environment variable to a dummy editor
	os.Setenv("EDITOR", "true")

	// Run the edit command
	cmd := exec.Command("go", "run", ".", "edit", clusterName)
	cmd.Env = append(os.Environ(), "EDITOR=true")
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Error running edit command: %v", err)
	}

	// Clean up
	err = os.Remove(configFile)
	if err != nil {
		t.Fatalf("Error removing config file: %v", err)
	}
	err = os.RemoveAll(configDir)
	if err != nil {
		t.Fatalf("Error removing config directory: %v", err)
	}
}
