package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSetDefaultCluster(t *testing.T) {
	// Point HOME at a temporary directory so the cluster config lives where
	// setDefaultCluster expects it.
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Create a sample cluster configuration file in the real config location.
	clusterName := "test-cluster"
	configFile := clusterConfigFile(clusterName)
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	if err := os.WriteFile(configFile, []byte("sample config"), 0644); err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}

	// Call the setDefaultCluster function
	err := setDefaultCluster(clusterName)
	if err != nil {
		t.Fatalf("Error setting default cluster: %v", err)
	}

	defaultConfigFile := filepath.Join(tempDir, ".kube", "config")

	// On platforms that support symlinks, verify the link target. Otherwise
	// (e.g. Windows without privileges) verify the file was copied.
	if runtime.GOOS != "windows" {
		linkTarget, err := os.Readlink(defaultConfigFile)
		if err != nil {
			t.Fatalf("Error reading symbolic link: %v", err)
		}
		if linkTarget != configFile {
			t.Errorf("Expected symbolic link target to be %s, but got %s", configFile, linkTarget)
		}
		return
	}

	content, err := os.ReadFile(defaultConfigFile)
	if err != nil {
		t.Fatalf("Error reading default config: %v", err)
	}
	if string(content) != "sample config" {
		t.Errorf("Expected default config content %q, but got %q", "sample config", string(content))
	}
}
