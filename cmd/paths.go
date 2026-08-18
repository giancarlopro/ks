package cmd

import (
	"os"
	"runtime"

	"github.com/giancarlopro/ks/config"
)

// clustersDir returns the directory where cluster configuration files live.
func clustersDir() string {
	return config.ClustersDir()
}

// clusterConfigFile returns the path to the configuration file for a cluster.
func clusterConfigFile(name string) string {
	return config.ClusterConfigFile(name)
}

// mergedConfigFile rebuilds every merged kubeconfig and returns the one for the
// given cluster. A merged kubeconfig holds the contexts of every registered
// cluster, so tools such as helmfile can select any of them by name.
func mergedConfigFile(name string) (string, error) {
	result, err := config.Rebuild(os.Stderr)
	if err != nil {
		return "", err
	}
	return result.Path(name)
}

// defaultShell returns the shell to spawn for an interactive session. It honors
// the user's SHELL on Unix and ComSpec on Windows, falling back to sensible
// per-platform defaults.
func defaultShell() string {
	if runtime.GOOS == "windows" {
		if shell := os.Getenv("ComSpec"); shell != "" {
			return shell
		}
		return "cmd.exe"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}

// defaultEditor returns the editor to open config files with, honoring the
// EDITOR environment variable and falling back per-platform.
func defaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vim"
}
