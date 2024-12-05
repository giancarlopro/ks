package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnterInteractiveShell(t *testing.T) {
	cluster := "test-cluster"
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, cluster+".yaml")

	cmd := exec.Command("zsh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("Error entering interactive shell: %v", err)
	}
}

func TestListClusters(t *testing.T) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	files, err := os.ReadDir(configDir)
	if err != nil {
		t.Errorf("Error listing clusters: %v", err)
	}

	var clusters []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		clusters = append(clusters, file.Name())
	}

	if len(clusters) == 0 {
		t.Errorf("No clusters found")
	}
}
