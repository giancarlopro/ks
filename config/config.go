package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ClusterConfig struct {
	APIVersion     string                 `yaml:"apiVersion"`
	Clusters       []Cluster              `yaml:"clusters"`
	Contexts       []Context              `yaml:"contexts"`
	CurrentContext string                 `yaml:"current-context"`
	Kind           string                 `yaml:"kind"`
	Preferences    map[string]interface{} `yaml:"preferences"`
	Users          []User                 `yaml:"users"`
}

type Cluster struct {
	Cluster ClusterDetails `yaml:"cluster"`
	Name    string         `yaml:"name"`
}

type ClusterDetails struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type Context struct {
	Context ContextDetails `yaml:"context"`
	Name    string         `yaml:"name"`
}

type ContextDetails struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	User      string `yaml:"user"`
}

type User struct {
	Name string     `yaml:"name"`
	User UserDetails `yaml:"user"`
}

type UserDetails struct {
	Exec ExecDetails `yaml:"exec"`
}

type ExecDetails struct {
	APIVersion        string      `yaml:"apiVersion"`
	Args              interface{} `yaml:"args"`
	Command           string      `yaml:"command"`
	Env               interface{} `yaml:"env"`
	InstallHint       string      `yaml:"installHint"`
	InteractiveMode   string      `yaml:"interactiveMode"`
	ProvideClusterInfo bool       `yaml:"provideClusterInfo"`
}

func ReadConfig(clusterName string) (*ClusterConfig, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, clusterName+".yaml")

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	var config ClusterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling configuration file: %w", err)
	}

	return &config, nil
}

func WriteConfig(config *ClusterConfig) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, config.Clusters[0].Name+".yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling configuration: %w", err)
	}

	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}

	return nil
}

func ValidateConfig(config *ClusterConfig) error {
	if config.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if len(config.Clusters) == 0 {
		return fmt.Errorf("at least one cluster is required")
	}
	for _, cluster := range config.Clusters {
		if cluster.Name == "" {
			return fmt.Errorf("cluster name is required")
		}
		if cluster.Cluster.Server == "" {
			return fmt.Errorf("server URL is required")
		}
	}
	if config.CurrentContext == "" {
		return fmt.Errorf("current-context is required")
	}
	if config.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if len(config.Users) == 0 {
		return fmt.Errorf("at least one user is required")
	}
	for _, user := range config.Users {
		if user.Name == "" {
			return fmt.Errorf("user name is required")
		}
		if user.User.Exec.Command == "" {
			return fmt.Errorf("exec command is required")
		}
	}
	return nil
}
