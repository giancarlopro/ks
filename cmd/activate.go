package cmd

import (
	"fmt"
	"os"
	"os/exec"

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
	if _, err := os.Stat(clusterConfigFile(cluster)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cluster %q does not exist", cluster)
		}
		return err
	}

	// KUBECONFIG points at the merged kubeconfig, not at the source file. The
	// merged file holds every registered context and activates this cluster.
	configFile, err := mergedConfigFile(cluster)
	if err != nil {
		return err
	}

	cmd := exec.Command(defaultShell())
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configFile))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
