package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [cluster-name]",
	Short: "Add a new cluster configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		configDir := filepath.Join(os.Getenv("HOME"), ".config", "ks", "clusters")

		// Create config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Println("Error creating config directory:", err)
			return
		}

		// Get editor from environment
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim" // Fallback to vim
		}

		// Create temporary file
		tmpfile, err := os.CreateTemp("", "ks-*.yaml")
		if err != nil {
			fmt.Println("Error creating temporary file:", err)
			return
		}
		defer os.Remove(tmpfile.Name())

		// Write template to temporary file
		template := `# Kubernetes cluster configuration
apiVersion: v1
clusters:
- cluster:
	server: https://
  name: ` + clusterName + `
`
		if _, err := tmpfile.Write([]byte(template)); err != nil {
			fmt.Println("Error writing template:", err)
			return
		}
		tmpfile.Close()

		// Open editor
		cmdEditor := exec.Command(editor, tmpfile.Name())
		cmdEditor.Stdin = os.Stdin
		cmdEditor.Stdout = os.Stdout
		cmdEditor.Stderr = os.Stderr

		if err := cmdEditor.Run(); err != nil {
			fmt.Println("Error opening editor:", err)
			return
		}

		// Read edited content
		content, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			fmt.Println("Error reading config:", err)
			return
		}

		// Save to final location
		configPath := filepath.Join(configDir, clusterName+".yaml")
		if err := os.WriteFile(configPath, content, 0600); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Println("Cluster registered successfully:", clusterName)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
