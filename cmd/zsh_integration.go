package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var zshIntegrationCmd = &cobra.Command{
	Use:   "zsh-integration",
	Short: "Set up Zsh and Oh-My-Zsh integration",
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current directory
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory:", err)
			return
		}

		// Check if .ksconfig file exists in the current directory
		ksconfigPath := filepath.Join(currentDir, ".ksconfig")
		if _, err := os.Stat(ksconfigPath); os.IsNotExist(err) {
			fmt.Println(".ksconfig file not found in the current directory")
			return
		}

		// Read the .ksconfig file
		content, err := os.ReadFile(ksconfigPath)
		if err != nil {
			fmt.Println("Error reading .ksconfig file:", err)
			return
		}
		cluster := strings.TrimSpace(string(content))

		if _, err := os.Stat(clusterConfigFile(cluster)); os.IsNotExist(err) {
			fmt.Println("Cluster configuration file not found:", cluster)
			return
		}

		// Set the KUBECONFIG environment variable to the merged kubeconfig.
		configFile, err := mergedConfigFile(cluster)
		if err != nil {
			fmt.Println("Error building merged config:", err)
			return
		}

		fmt.Printf("Setting KUBECONFIG to %s\n", configFile)
		os.Setenv("KUBECONFIG", configFile)
	},
}

func init() {
	rootCmd.AddCommand(zshIntegrationCmd)
}
