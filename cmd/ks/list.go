package ks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered Kubernetes clusters",
	Run: func(cmd *cobra.Command, args []string) {
		clusters, err := listClusters()
		if err != nil {
			fmt.Println("Error listing clusters:", err)
			return
		}

		for _, cluster := range clusters {
			fmt.Println(cluster)
		}
	},
}

func listClusters() ([]string, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	files, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var clusters []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		clusters = append(clusters, file.Name())
	}

	return clusters, nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
