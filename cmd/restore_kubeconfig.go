package cmd

import (
	"fmt"
	"os"

	"github.com/giancarlopro/ks/config"
	"github.com/spf13/cobra"
)

var restoreForce bool

var restoreKubeconfigCmd = &cobra.Command{
	Use:   "restore-kubeconfig",
	Short: "Restore the kubeconfig that ks import moved aside",
	Long: "Restore the kubeconfig that `ks import` moved aside.\n\n" +
		"It moves ~/.kube/config.bkp back to ~/.kube/config. The clusters that\n" +
		"import registered stay registered.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return restoreKubeconfig()
	},
}

func init() {
	restoreKubeconfigCmd.Flags().BoolVar(&restoreForce, "force", false, "overwrite a default config that ks did not create")
	rootCmd.AddCommand(restoreKubeconfigCmd)
}

func restoreKubeconfig() error {
	defaultConfigFile := config.DefaultConfigFile()
	backupFile := config.KubeBackupFile()

	if _, err := os.Lstat(backupFile); err != nil {
		return fmt.Errorf("no backup at %s", backupFile)
	}

	// Only a file ks created is safe to remove. Anything else is the user's
	// own work.
	if _, err := os.Lstat(defaultConfigFile); err == nil {
		if !config.Managed(defaultConfigFile) && !restoreForce {
			return fmt.Errorf("%s is not a file ks created; pass --force to overwrite it", defaultConfigFile)
		}
		if err := os.Remove(defaultConfigFile); err != nil {
			return fmt.Errorf("failed to remove %s: %w", defaultConfigFile, err)
		}
	}

	if err := os.Rename(backupFile, defaultConfigFile); err != nil {
		return fmt.Errorf("failed to restore %s: %w", defaultConfigFile, err)
	}

	// The default config is no longer a ks file, so the record of the default
	// cluster no longer describes it.
	if err := os.Remove(config.DefaultRecordFile()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear the default cluster record: %w", err)
	}

	fmt.Printf("Restored %s from %s.\n", defaultConfigFile, backupFile)

	return nil
}
