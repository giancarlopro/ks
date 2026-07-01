package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	configFile := clusterConfigFile(cluster)
	kubeDir := filepath.Join(config.HomeDir(), ".kube")
	defaultConfigFile := filepath.Join(kubeDir, "config")

	// Make sure the cluster we're pointing at actually exists.
	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cluster %q does not exist", cluster)
		}
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
		if err := copyFile(configFile, defaultConfigFile); err != nil {
			return fmt.Errorf("failed to copy config file: %w", err)
		}
	}

	return nil
}

// copyFile copies the contents of src to dst, creating or truncating dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
