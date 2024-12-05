package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ClusterConfig struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	Token  string `yaml:"token"`
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
	configFile := filepath.Join(configDir, config.Name+".yaml")

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
	if config.Name == "" {
		return fmt.Errorf("cluster name is required")
	}
	if config.Server == "" {
		return fmt.Errorf("server URL is required")
	}
	if config.Token == "" {
		return fmt.Errorf("authentication token is required")
	}
	return nil
}
