package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

		// Create the initial kubectl format configuration
		initialConfig := map[string]interface{}{
			"apiVersion": "v1",
			"clusters": []map[string]interface{}{
				{
					"cluster": map[string]interface{}{
						"certificate-authority-data": "",
						"server":                     "",
					},
					"name": clusterName,
				},
			},
			"contexts": []map[string]interface{}{
				{
					"context": map[string]interface{}{
						"cluster":   clusterName,
						"namespace": "default",
						"user":      clusterName,
					},
					"name": clusterName,
				},
			},
			"current-context": clusterName,
			"kind":            "Config",
			"preferences":     map[string]interface{}{},
			"users": []map[string]interface{}{
				{
					"name": clusterName,
					"user": map[string]interface{}{
						"exec": map[string]interface{}{
							"apiVersion":        "client.authentication.k8s.io/v1beta1",
							"args":              nil,
							"command":           "gke-gcloud-auth-plugin",
							"env":               nil,
							"installHint":       "Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin",
							"interactiveMode":   "IfAvailable",
							"provideClusterInfo": true,
						},
					},
				},
			},
		}

		// Write the initial configuration to the file
		data, err := yaml.Marshal(&initialConfig)
		if err != nil {
			fmt.Println("Error marshalling initial configuration:", err)
			return
		}

		if _, err := file.Write(data); err != nil {
			fmt.Println("Error writing initial configuration to file:", err)
			return
		}

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
