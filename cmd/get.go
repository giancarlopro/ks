package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var getCmd = &cobra.Command{
	Use:   "get <cluster-name>",
	Short: "Get the details of a specific Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		clusterDetails, err := getClusterDetails(clusterName)
		if err != nil {
			fmt.Println("Error getting cluster details:", err)
			return
		}

		fmt.Println(clusterDetails)
	},
}

func getClusterDetails(clusterName string) (string, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, clusterName)

	file, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	var clusterDetails map[string]interface{}
	err = yaml.Unmarshal(file, &clusterDetails)
	if err != nil {
		return "", err
	}

	clusterDetailsStr, err := yaml.Marshal(&clusterDetails)
	if err != nil {
		return "", err
	}

	return string(clusterDetailsStr), nil
}

func init() {
	rootCmd.AddCommand(getCmd)
}
