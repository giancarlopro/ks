package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateBackup(t *testing.T) {
	clusterName := "test-cluster"
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, clusterName+".yaml")
	backupDir := filepath.Join(configDir, "backups")
	backupFile := filepath.Join(backupDir, clusterName+".yaml.bak")

	// Create a test configuration file
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Error creating config directory: %v", err)
	}
	err = ioutil.WriteFile(configFile, []byte("test config data"), 0644)
	if err != nil {
		t.Fatalf("Error creating config file: %v", err)
	}

	// Create a backup
	err = CreateBackup(clusterName)
	if err != nil {
		t.Fatalf("Error creating backup: %v", err)
	}

	// Check if the backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Fatalf("Backup file does not exist: %s", backupFile)
	}

	// Clean up
	os.RemoveAll(configDir)
}

func TestRecoverFromBackup(t *testing.T) {
	clusterName := "test-cluster"
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, clusterName+".yaml")
	backupDir := filepath.Join(configDir, "backups")
	backupFile := filepath.Join(backupDir, clusterName+".yaml.bak")

	// Create a test backup file
	err := os.MkdirAll(backupDir, 0755)
	if err != nil {
		t.Fatalf("Error creating backup directory: %v", err)
	}
	err = ioutil.WriteFile(backupFile, []byte("test backup data"), 0644)
	if err != nil {
		t.Fatalf("Error creating backup file: %v", err)
	}

	// Recover from the backup
	err = RecoverFromBackup(clusterName)
	if err != nil {
		t.Fatalf("Error recovering from backup: %v", err)
	}

	// Check if the configuration file exists and has the correct content
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Error reading config file: %v", err)
	}
	if string(data) != "test backup data" {
		t.Fatalf("Config file content does not match backup data")
	}

	// Clean up
	os.RemoveAll(configDir)
}
