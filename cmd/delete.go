package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <cluster-name>",
	Short: "Delete a registered Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		configFile := clusterConfigFile(clusterName)

		// Check if the cluster exists
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Println("Cluster does not exist:", clusterName)
			return
		}

		// Delete the configuration file
		if err := os.Remove(configFile); err != nil {
			fmt.Println("Error deleting configuration file:", err)
			return
		}

		fmt.Println("Cluster deleted successfully:", clusterName)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
