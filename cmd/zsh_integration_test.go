package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/giancarlopro/ks/config"
)

func TestZshIntegration(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// A .ksconfig file names the cluster of a project directory. echo adds a
	// newline, so the command must trim the name.
	clusterName := "test-cluster"
	ksconfigPath := filepath.Join(tempDir, ".ksconfig")
	if err := os.WriteFile(ksconfigPath, []byte(clusterName+"\n"), 0644); err != nil {
		t.Fatalf("Error creating .ksconfig file: %v", err)
	}

	registerCluster(t, clusterName)

	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error reading working directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Error changing directory: %v", err)
	}
	defer os.Chdir(workingDir)

	zshIntegrationCmd.Run(zshIntegrationCmd, []string{})

	// KUBECONFIG points at the merged kubeconfig, not at the source file.
	expected := config.GeneratedConfigFile(clusterName)
	if actual := os.Getenv("KUBECONFIG"); actual != expected {
		t.Errorf("KUBECONFIG environment variable not set correctly. Expected: %s, Got: %s", expected, actual)
	}
}
