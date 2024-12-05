package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ks",
	Short: "ks is a CLI tool for managing Kubernetes clusters",
	Run: func(cmd *cobra.Command, args []string) {
		// Display the list of clusters
		clusters, err := listClusters()
		if err != nil {
			fmt.Println("Error listing clusters:", err)
			return
		}

		// Use promptui to select a cluster
		prompt := promptui.Select{
			Label: "Select a cluster",
			Items: clusters,
		}

		_, cluster, err := prompt.Run()
		if err != nil {
			fmt.Println("Error selecting cluster:", err)
			return
		}

		// Enter an interactive shell with the selected cluster
		err = enterInteractiveShell(cluster)
		if err != nil {
			fmt.Println("Error entering interactive shell:", err)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func listClusters() ([]string, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var clusters []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		clusters = append(clusters, file.Name())
	}

	return clusters, nil
}

func enterInteractiveShell(cluster string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, cluster)

	cmd := exec.Command("zsh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
