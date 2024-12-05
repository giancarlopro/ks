package ks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <cluster-name>",
	Short: "Edit the configuration of a registered Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
		configFile := filepath.Join(configDir, clusterName+".yaml")

		// Check if the cluster exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Println("Cluster does not exist:", clusterName)
			return
		}

		// Open the default editor for the user to edit the configuration file
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}

		cmdEditor := exec.Command(editor, configFile)
		cmdEditor.Stdin = os.Stdin
		cmdEditor.Stdout = os.Stdout
		cmdEditor.Stderr = os.Stderr

		if err := cmdEditor.Run(); err != nil {
			fmt.Println("Error opening editor:", err)
			return
		}

		fmt.Println("Cluster configuration updated successfully:", clusterName)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
