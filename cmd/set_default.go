package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/giancarlopro/ks/config"
	"github.com/spf13/cobra"
)

var setDefaultCmd = &cobra.Command{
	Use:   "set-default <cluster-name>",
	Short: "Set a default Kubernetes cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		err := setDefaultCluster(cluster)
		if err != nil {
			fmt.Println("Error setting default cluster:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}

func setDefaultCluster(cluster string) error {
	kubeDir := config.KubeDir()
	defaultConfigFile := config.DefaultConfigFile()

	// Make sure the cluster we're pointing at actually exists.
	if _, err := os.Stat(clusterConfigFile(cluster)); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cluster %q does not exist", cluster)
		}
		return err
	}

	// Record the choice before rebuilding, so a later rebuild can refresh the
	// default config on systems that copy it instead of linking it.
	if err := config.WriteDefaultRecord(cluster); err != nil {
		return err
	}

	// Point the default config at the merged kubeconfig, so tools outside a ks
	// shell see every registered context.
	configFile, err := mergedConfigFile(cluster)
	if err != nil {
		return err
	}

	// Ensure the ~/.kube directory exists.
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		return fmt.Errorf("failed to create kube directory: %w", err)
	}

	// Remove the existing config (symlink or file) if it exists.
	if _, err := os.Lstat(defaultConfigFile); err == nil {
		if err := os.Remove(defaultConfigFile); err != nil {
			return fmt.Errorf("failed to remove existing default config: %w", err)
		}
	}

	// Create a new symbolic link. Windows often disallows symlinks without
	// elevated privileges or developer mode, so fall back to copying the file.
	if err := os.Symlink(configFile, defaultConfigFile); err != nil {
		if runtime.GOOS != "windows" {
			return fmt.Errorf("failed to create symbolic link: %w", err)
		}
		if err := config.CopyFile(configFile, defaultConfigFile); err != nil {
			return fmt.Errorf("failed to copy config file: %w", err)
		}
	}

	return nil
}
