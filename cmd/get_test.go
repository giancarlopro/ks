package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetClusterDetails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	clusterName := "test-cluster"
	configFile := clusterConfigFile(clusterName)

	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}

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
	if err := os.WriteFile(configFile, []byte(testData), 0600); err != nil {
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
	// getClusterDetails renders the file, so it emits no leading newline.
	expectedDetails = strings.TrimPrefix(expectedDetails, "\n")

	if clusterDetails != expectedDetails {
		t.Errorf("Cluster details do not match. Expected: %s, Got: %s", expectedDetails, clusterDetails)
	}
}
