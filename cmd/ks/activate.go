package ks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate <cluster-name>",
	Short: "Activate a Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		err := activateCluster(cluster)
		if err != nil {
			fmt.Println("Error activating cluster:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(activateCmd)
}

func activateCluster(cluster string) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")
	configFile := filepath.Join(configDir, cluster+".yaml")

	cmd := exec.Command("zsh")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
