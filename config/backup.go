package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CreateBackup(clusterName string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, clusterName+".yaml")
	backupDir := filepath.Join(configDir, "backups")

	// Create the backup directory if it doesn't exist
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("error creating backup directory: %w", err)
		}
	}

	backupFile := filepath.Join(backupDir, clusterName+".yaml.bak")

	// Read the existing configuration file
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading configuration file: %w", err)
	}

	// Write the backup file
	if err := ioutil.WriteFile(backupFile, data, 0644); err != nil {
		return fmt.Errorf("error writing backup file: %w", err)
	}

	return nil
}

func RecoverFromBackup(clusterName string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, clusterName+".yaml")
	backupDir := filepath.Join(configDir, "backups")
	backupFile := filepath.Join(backupDir, clusterName+".yaml.bak")

	// Check if the backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	// Read the backup file
	data, err := ioutil.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("error reading backup file: %w", err)
	}

	// Write the configuration file
	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}

	return nil
}
