package config

import (
	"os"
	"path/filepath"
)

// HomeDir returns the current user's home directory in a cross-platform way.
// It relies on os.UserHomeDir (which honors HOME on Unix and USERPROFILE on
// Windows) and falls back to the HOME environment variable if that fails.
func HomeDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	return os.Getenv("HOME")
}

// ClustersDir returns the directory where cluster configuration files live.
func ClustersDir() string {
	return filepath.Join(HomeDir(), ".config", "ks", "clusters")
}

// ClusterConfigFile returns the path to the configuration file for a cluster.
func ClusterConfigFile(name string) string {
	return filepath.Join(ClustersDir(), name+".yaml")
}

// GeneratedDir returns the directory where merged kubeconfigs live. ks writes
// these files; the user never edits them.
func GeneratedDir() string {
	return filepath.Join(HomeDir(), ".config", "ks", "generated")
}

// GeneratedConfigFile returns the path to the merged kubeconfig for a cluster.
func GeneratedConfigFile(name string) string {
	return filepath.Join(GeneratedDir(), name+".yaml")
}

// DefaultRecordFile returns the path of the file that records the cluster set
// with `ks set-default`.
func DefaultRecordFile() string {
	return filepath.Join(HomeDir(), ".config", "ks", "default")
}

// KubeDir returns the directory holding the default kubectl configuration.
func KubeDir() string {
	return filepath.Join(HomeDir(), ".kube")
}

// DefaultConfigFile returns the path of the default kubectl configuration.
func DefaultConfigFile() string {
	return filepath.Join(KubeDir(), "config")
}
