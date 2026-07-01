package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <cluster-name>",
	Short: "Edit the configuration of a registered Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		configFile := clusterConfigFile(clusterName)

		// Check if the cluster exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Println("Cluster does not exist:", clusterName)
			return
		}

		// Open the default editor for the user to edit the configuration file
		cmdEditor := exec.Command(defaultEditor(), configFile)
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
