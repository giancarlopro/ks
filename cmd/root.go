package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

		for i, cluster := range clusters {
			fmt.Printf("%d. %s\n", i+1, cluster)
		}

		// Prompt the user to select a cluster
		var choice int
		fmt.Print("Select a cluster: ")
		fmt.Scan(&choice)

		if choice < 1 || choice > len(clusters) {
			fmt.Println("Invalid choice")
			return
		}

		// Enter an interactive shell with the selected cluster
		cluster := clusters[choice-1]
		err = enterInteractiveShell(cluster)
		if err != nil {
			fmt.Println("Error entering interactive shell:", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func enterInteractiveShell(cluster string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, cluster+".yaml")

	cmd := exec.Command("zsh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
