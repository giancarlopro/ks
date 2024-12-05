package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListClusters(t *testing.T) {
	// Set up test environment
	homeDir := os.Getenv("HOME")
	configDir := filepath.Join(homeDir, ".config", "ks", "clusters")
	clusterName1 := "test-cluster-1"
	clusterName2 := "test-cluster-2"
	configFile1 := filepath.Join(configDir, clusterName1+".yaml")
	configFile2 := filepath.Join(configDir, clusterName2+".yaml")

	// Create temporary configuration files
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	defer os.RemoveAll(configDir)

	file1, err := os.Create(configFile1)
	if err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}
	defer file1.Close()

	file2, err := os.Create(configFile2)
	if err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}
	defer file2.Close()

	// Run the listClusters function
	clusters, err := listClusters()
	if err != nil {
		t.Fatalf("Error listing clusters: %v", err)
	}

	// Verify the clusters
	expectedClusters := []string{clusterName1 + ".yaml", clusterName2 + ".yaml"}
	if len(clusters) != len(expectedClusters) {
		t.Errorf("Expected %d clusters, got %d", len(expectedClusters), len(clusters))
	}

	for i, cluster := range clusters {
		if cluster != expectedClusters[i] {
			t.Errorf("Expected cluster %s, got %s", expectedClusters[i], cluster)
		}
	}
}
