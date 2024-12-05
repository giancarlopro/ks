package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestReadConfig(t *testing.T) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, "test-cluster.yaml")

	// Create the test configuration directory if it doesn't exist
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}

	// Create a test configuration file
	initialConfig := ClusterConfig{
		APIVersion: "v1",
		Clusters: []Cluster{
			{
				Cluster: ClusterDetails{
					CertificateAuthorityData: "test-ca",
					Server:                   "https://test-server",
				},
				Name: "test-cluster",
			},
		},
		Contexts: []Context{
			{
				Context: ContextDetails{
					Cluster:   "test-cluster",
					Namespace: "default",
					User:      "test-user",
				},
				Name: "test-context",
			},
		},
		CurrentContext: "test-context",
		Kind:           "Config",
		Preferences:    map[string]interface{}{},
		Users: []User{
			{
				Name: "test-user",
				User: UserDetails{
					Exec: ExecDetails{
						APIVersion:        "client.authentication.k8s.io/v1beta1",
						Args:              nil,
						Command:           "test-command",
						Env:               nil,
						InstallHint:       "test-install-hint",
						InteractiveMode:   "IfAvailable",
						ProvideClusterInfo: true,
					},
				},
			},
		},
	}

	data, err := yaml.Marshal(&initialConfig)
	if err != nil {
		t.Fatalf("Error marshalling initial configuration: %v", err)
	}

	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("Error writing test configuration file: %v", err)
	}
	defer os.Remove(configFile)

	// Read the configuration file
	config, err := ReadConfig("test-cluster")
	if err != nil {
		t.Fatalf("Error reading configuration file: %v", err)
	}

	// Validate the configuration
	if config.APIVersion != initialConfig.APIVersion {
		t.Errorf("Expected apiVersion %s, got %s", initialConfig.APIVersion, config.APIVersion)
	}
	if len(config.Clusters) != len(initialConfig.Clusters) {
		t.Errorf("Expected %d clusters, got %d", len(initialConfig.Clusters), len(config.Clusters))
	}
	if config.Clusters[0].Name != initialConfig.Clusters[0].Name {
		t.Errorf("Expected cluster name %s, got %s", initialConfig.Clusters[0].Name, config.Clusters[0].Name)
	}
	if config.Clusters[0].Cluster.Server != initialConfig.Clusters[0].Cluster.Server {
		t.Errorf("Expected server %s, got %s", initialConfig.Clusters[0].Cluster.Server, config.Clusters[0].Cluster.Server)
	}
	if config.CurrentContext != initialConfig.CurrentContext {
		t.Errorf("Expected current-context %s, got %s", initialConfig.CurrentContext, config.CurrentContext)
	}
	if config.Kind != initialConfig.Kind {
		t.Errorf("Expected kind %s, got %s", initialConfig.Kind, config.Kind)
	}
	if len(config.Users) != len(initialConfig.Users) {
		t.Errorf("Expected %d users, got %d", len(initialConfig.Users), len(config.Users))
	}
	if config.Users[0].Name != initialConfig.Users[0].Name {
		t.Errorf("Expected user name %s, got %s", initialConfig.Users[0].Name, config.Users[0].Name)
	}
	if config.Users[0].User.Exec.Command != initialConfig.Users[0].User.Exec.Command {
		t.Errorf("Expected exec command %s, got %s", initialConfig.Users[0].User.Exec.Command, config.Users[0].User.Exec.Command)
	}
}

func TestWriteConfig(t *testing.T) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, "test-cluster.yaml")

	// Create the test configuration directory if it doesn't exist
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}

	// Create a test configuration
	initialConfig := ClusterConfig{
		APIVersion: "v1",
		Clusters: []Cluster{
			{
				Cluster: ClusterDetails{
					CertificateAuthorityData: "test-ca",
					Server:                   "https://test-server",
				},
				Name: "test-cluster",
			},
		},
		Contexts: []Context{
			{
				Context: ContextDetails{
					Cluster:   "test-cluster",
					Namespace: "default",
					User:      "test-user",
				},
				Name: "test-context",
			},
		},
		CurrentContext: "test-context",
		Kind:           "Config",
		Preferences:    map[string]interface{}{},
		Users: []User{
			{
				Name: "test-user",
				User: UserDetails{
					Exec: ExecDetails{
						APIVersion:        "client.authentication.k8s.io/v1beta1",
						Args:              nil,
						Command:           "test-command",
						Env:               nil,
						InstallHint:       "test-install-hint",
						InteractiveMode:   "IfAvailable",
						ProvideClusterInfo: true,
					},
				},
			},
		},
	}

	// Write the configuration file
	if err := WriteConfig(&initialConfig); err != nil {
		t.Fatalf("Error writing configuration file: %v", err)
	}
	defer os.Remove(configFile)

	// Read the configuration file
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Error reading configuration file: %v", err)
	}

	var config ClusterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatalf("Error unmarshalling configuration file: %v", err)
	}

	// Validate the configuration
	if config.APIVersion != initialConfig.APIVersion {
		t.Errorf("Expected apiVersion %s, got %s", initialConfig.APIVersion, config.APIVersion)
	}
	if len(config.Clusters) != len(initialConfig.Clusters) {
		t.Errorf("Expected %d clusters, got %d", len(initialConfig.Clusters), len(config.Clusters))
	}
	if config.Clusters[0].Name != initialConfig.Clusters[0].Name {
		t.Errorf("Expected cluster name %s, got %s", initialConfig.Clusters[0].Name, config.Clusters[0].Name)
	}
	if config.Clusters[0].Cluster.Server != initialConfig.Clusters[0].Cluster.Server {
		t.Errorf("Expected server %s, got %s", initialConfig.Clusters[0].Cluster.Server, config.Clusters[0].Cluster.Server)
	}
	if config.CurrentContext != initialConfig.CurrentContext {
		t.Errorf("Expected current-context %s, got %s", initialConfig.CurrentContext, config.CurrentContext)
	}
	if config.Kind != initialConfig.Kind {
		t.Errorf("Expected kind %s, got %s", initialConfig.Kind, config.Kind)
	}
	if len(config.Users) != len(initialConfig.Users) {
		t.Errorf("Expected %d users, got %d", len(initialConfig.Users), len(config.Users))
	}
	if config.Users[0].Name != initialConfig.Users[0].Name {
		t.Errorf("Expected user name %s, got %s", initialConfig.Users[0].Name, config.Users[0].Name)
	}
	if config.Users[0].User.Exec.Command != initialConfig.Users[0].User.Exec.Command {
		t.Errorf("Expected exec command %s, got %s", initialConfig.Users[0].User.Exec.Command, config.Users[0].User.Exec.Command)
	}
}

func TestValidateConfig(t *testing.T) {
	validConfig := ClusterConfig{
		APIVersion: "v1",
		Clusters: []Cluster{
			{
				Cluster: ClusterDetails{
					CertificateAuthorityData: "test-ca",
					Server:                   "https://test-server",
				},
				Name: "test-cluster",
			},
		},
		Contexts: []Context{
			{
				Context: ContextDetails{
					Cluster:   "test-cluster",
					Namespace: "default",
					User:      "test-user",
				},
				Name: "test-context",
			},
		},
		CurrentContext: "test-context",
		Kind:           "Config",
		Preferences:    map[string]interface{}{},
		Users: []User{
			{
				Name: "test-user",
				User: UserDetails{
					Exec: ExecDetails{
						APIVersion:        "client.authentication.k8s.io/v1beta1",
						Args:              nil,
						Command:           "test-command",
						Env:               nil,
						InstallHint:       "test-install-hint",
						InteractiveMode:   "IfAvailable",
						ProvideClusterInfo: true,
					},
				},
			},
		},
	}

	if err := ValidateConfig(&validConfig); err != nil {
		t.Errorf("Expected valid configuration, got error: %v", err)
	}

	invalidConfig := ClusterConfig{
		APIVersion: "",
		Clusters:   []Cluster{},
		Contexts:   []Context{},
		Users:      []User{},
	}

	if err := ValidateConfig(&invalidConfig); err == nil {
		t.Errorf("Expected invalid configuration, got no error")
	}
}
