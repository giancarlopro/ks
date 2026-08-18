package cmd

import (
	"os"
	"runtime"
	"testing"

	"github.com/giancarlopro/ks/config"
)

func TestSetDefaultCluster(t *testing.T) {
	// Point HOME at a temporary directory so the cluster config lives where
	// setDefaultCluster expects it.
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	clusterName := "test-cluster"
	registerCluster(t, clusterName)

	if err := setDefaultCluster(clusterName); err != nil {
		t.Fatalf("Error setting default cluster: %v", err)
	}

	// The default config points at the merged kubeconfig, so tools outside a
	// ks shell see every registered context.
	mergedFile := config.GeneratedConfigFile(clusterName)
	defaultConfigFile := config.DefaultConfigFile()

	if got := currentContextOf(t, mergedFile); got != clusterName {
		t.Errorf("Expected current-context %s, got %q", clusterName, got)
	}

	// set-default records the choice, so a later rebuild can refresh a copied
	// default config.
	record, err := os.ReadFile(config.DefaultRecordFile())
	if err != nil {
		t.Fatalf("Error reading default record: %v", err)
	}
	if string(record) != clusterName+"\n" {
		t.Errorf("Expected the record to name %s, got %q", clusterName, string(record))
	}

	// On platforms that support symlinks, verify the link target. Otherwise
	// (e.g. Windows without privileges) verify the file was copied.
	if runtime.GOOS != "windows" {
		linkTarget, err := os.Readlink(defaultConfigFile)
		if err != nil {
			t.Fatalf("Error reading symbolic link: %v", err)
		}
		if linkTarget != mergedFile {
			t.Errorf("Expected symbolic link target to be %s, but got %s", mergedFile, linkTarget)
		}
		return
	}

	content, err := os.ReadFile(defaultConfigFile)
	if err != nil {
		t.Fatalf("Error reading default config: %v", err)
	}
	merged, err := os.ReadFile(mergedFile)
	if err != nil {
		t.Fatalf("Error reading merged config: %v", err)
	}
	if string(content) != string(merged) {
		t.Error("Expected the default config to hold the merged config")
	}
}

func TestSetDefaultClusterUnknown(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := setDefaultCluster("missing"); err == nil {
		t.Error("Expected an error for a cluster that does not exist")
	}
}
