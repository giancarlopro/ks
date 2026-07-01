package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ks [cluster-name]",
	Short: "ks is a CLI tool for managing Kubernetes clusters",
	Long: "ks is a CLI tool for managing Kubernetes clusters.\n\n" +
		"Run `ks` with no arguments to pick a cluster interactively, or\n" +
		"`ks <cluster-name>` to activate a cluster directly.",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If a cluster name is passed directly, activate it without prompting.
		if len(args) == 1 {
			if err := activateCluster(args[0]); err != nil {
				fmt.Println("Error activating cluster:", err)
			}
			return
		}

		// Otherwise, display the list of clusters to choose from.
		clusters, err := listClusters()
		if err != nil {
			fmt.Println("Error listing clusters:", err)
			return
		}

		if len(clusters) == 0 {
			fmt.Println("No clusters found")
			return
		}

		selectedCluster, err := selectCluster(clusters)
		if err != nil {
			fmt.Println("Error selecting cluster:", err)
			return
		}

		if err := activateCluster(selectedCluster); err != nil {
			fmt.Println("Error activating cluster:", err)
		}
	},
}

// selectCluster prompts the user to choose a cluster, using fzf when available
// and falling back to a simple prompt otherwise.
func selectCluster(clusters []string) (string, error) {
	if isFzfAvailable() {
		idx, err := fuzzyfinder.Find(clusters, func(i int) string {
			return clusters[i]
		})
		if err != nil {
			return "", fmt.Errorf("error using fzf: %w", err)
		}
		return clusters[idx], nil
	}

	prompt := promptui.Select{
		Label: "Select a cluster",
		Items: clusters,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("error using prompt: %w", err)
	}
	return result, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func isFzfAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}
