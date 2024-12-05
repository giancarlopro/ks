package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetClusterDetails(t *testing.T) {
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

	// Write test data to the configuration file
	testData := `
apiVersion: v1
clusters:
  - cluster:
      certificate-authority-data: test-ca
      server: test-server
    name: test-cluster
contexts:
  - context:
      cluster: test-cluster
      namespace: test-namespace
      user: test-user
    name: test-cluster
current-context: test-cluster
kind: Config
preferences: {}
users:
  - name: test-user
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args: null
        command: test-command
        env: null
        installHint: test-installHint
        interactiveMode: IfAvailable
        provideClusterInfo: true
`
	_, err = file.WriteString(testData)
	if err != nil {
		t.Fatalf("Error writing test data to config file: %v", err)
	}

	// Run the getClusterDetails function
	clusterDetails, err := getClusterDetails(clusterName)
	if err != nil {
		t.Errorf("Error getting cluster details: %v", err)
	}

	// Verify the cluster details
	expectedDetails := `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: test-ca
    server: test-server
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    namespace: test-namespace
    user: test-user
  name: test-cluster
current-context: test-cluster
kind: Config
preferences: {}
users:
- name: test-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args: null
      command: test-command
      env: null
      installHint: test-installHint
      interactiveMode: IfAvailable
      provideClusterInfo: true
`
	if clusterDetails != expectedDetails {
		t.Errorf("Cluster details do not match. Expected: %s, Got: %s", expectedDetails, clusterDetails)
	}
}
