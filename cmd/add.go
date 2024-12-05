package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <cluster-name>",
	Short: "Register a new Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
		configFile := filepath.Join(configDir, clusterName+".yaml")

		// Check if the cluster already exists
		if _, err := os.Stat(configFile); err == nil {
			fmt.Println("Cluster already exists:", clusterName)
			return
		}

		// Create the configuration file
		file, err := os.Create(configFile)
		if err != nil {
			fmt.Println("Error creating configuration file:", err)
			return
		}
		defer file.Close()

		// Open the default editor for the user to set the content of the file
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

		fmt.Println("Cluster registered successfully:", clusterName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
