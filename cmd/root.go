package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ktr0731/go-fuzzyfinder"
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

		if len(clusters) == 0 {
			fmt.Println("No clusters found")
			return
		}

		var selectedCluster string
		if isFzfAvailable() {
			idx, err := fuzzyfinder.Find(
				clusters,
				func(i int) string {
					return clusters[i]
				},
			)
			if err != nil {
				fmt.Println("Error using fzf:", err)
				return
			}
			selectedCluster = clusters[idx]
		} else {
			prompt := promptui.Select{
				Label: "Select a cluster",
				Items: clusters,
			}
			_, result, err := prompt.Run()
			if err != nil {
				fmt.Println("Error using promptui:", err)
				return
			}
			selectedCluster = result
		}

		// Enter an interactive shell with the selected cluster
		err = enterInteractiveShell(selectedCluster)
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

func isFzfAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}
