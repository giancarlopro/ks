package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var setDefaultCmd = &cobra.Command{
	Use:   "set-default <cluster-name>",
	Short: "Set a default Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		err := setDefaultCluster(cluster)
		if err != nil {
			fmt.Println("Error setting default cluster:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}

func setDefaultCluster(cluster string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, cluster)
	defaultConfigFile := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	// Remove existing symbolic link if it exists
	if _, err := os.Lstat(defaultConfigFile); err == nil {
		if err := os.Remove(defaultConfigFile); err != nil {
			return fmt.Errorf("failed to remove existing symbolic link: %w", err)
		}
	}

	// Create a new symbolic link
	if err := os.Symlink(configFile, defaultConfigFile); err != nil {
		return fmt.Errorf("failed to create symbolic link: %w", err)
	}

	return nil
}
