package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/giancarlopro/ks/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// promptIn is where the prompts read their answers. A test replaces it.
var promptIn io.Reader = os.Stdin

var (
	importAs  string
	importYes bool
)

var importCmd = &cobra.Command{
	Use:   "import [path...]",
	Short: "Import a kubeconfig that ks does not manage",
	Long: "Import a kubeconfig that ks does not manage.\n\n" +
		"Import registers one cluster for each context, and asks for a short\n" +
		"name for each of them. It then moves the default kubectl config to\n" +
		"~/.kube/config.bkp, and points it at ks. Undo that move with\n" +
		"`ks restore-kubeconfig`.\n\n" +
		"With no path, import reads KUBECONFIG when it is set, and\n" +
		"~/.kube/config otherwise.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return importKubeconfig(args)
	},
}

func init() {
	importCmd.Flags().StringVar(&importAs, "as", "", "import every context into one cluster with this name")
	importCmd.Flags().BoolVarP(&importYes, "yes", "y", false, "accept every suggested name, and move the default config without asking")
	rootCmd.AddCommand(importCmd)
}

// registration is one cluster that import is about to write.
type registration struct {
	name    string
	context config.SourceContext
	doc     yaml.MapSlice
}

func importKubeconfig(args []string) error {
	paths, err := config.SourcePaths(args)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(promptIn)

	registrations, err := planImport(reader, paths)
	if err != nil {
		return err
	}
	// Every context was skipped. That is a choice, not a fault.
	if len(registrations) == 0 {
		fmt.Println("Nothing registered.")
		return nil
	}

	// Every name is chosen before anything is written, so a cancelled prompt
	// leaves no half-imported state.
	for _, r := range registrations {
		if err := config.WriteCluster(r.name, r.doc); err != nil {
			return err
		}
	}

	names := make([]string, 0, len(registrations))
	for _, r := range registrations {
		names = append(names, r.name)
	}
	fmt.Printf("Registered %d cluster(s): %s\n", len(names), strings.Join(names, ", "))

	return takeOverDefaultConfig(reader, activeClusterName(registrations))
}

// planImport reads the source kubeconfigs and picks a name for each cluster it
// will register.
func planImport(reader *bufio.Reader, paths []string) ([]registration, error) {
	if importAs != "" {
		doc, err := config.ReadSourcesAsOne(paths)
		if err != nil {
			return nil, err
		}
		if config.Registered(importAs) {
			return nil, fmt.Errorf("cluster %q already exists", importAs)
		}
		return []registration{{name: importAs, doc: doc}}, nil
	}

	contexts, warnings, err := config.ReadSources(paths)
	if err != nil {
		return nil, err
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}
	if len(contexts) == 0 {
		return nil, fmt.Errorf("no context to import")
	}

	var registrations []registration
	taken := map[string]bool{}

	for _, ctx := range contexts {
		name, err := chooseName(reader, ctx, taken)
		if err != nil {
			return nil, err
		}
		if name == "" {
			continue
		}
		taken[name] = true
		registrations = append(registrations, registration{name: name, context: ctx, doc: ctx.Config()})
	}

	return registrations, nil
}

// chooseName asks for the name to register one context under. It returns an
// empty name when the user skips the context.
func chooseName(reader *bufio.Reader, ctx config.SourceContext, taken map[string]bool) (string, error) {
	if importYes {
		return config.FreeName(ctx.Suggested, taken), nil
	}

	fmt.Printf("\n%s\n  server: %s\n", ctx.Context, ctx.Server)

	for {
		answer, err := ask(reader, fmt.Sprintf("Name [%s] (Enter accepts, - skips): ", ctx.Suggested))
		if err != nil {
			return "", fmt.Errorf("import cancelled: %w", err)
		}

		if answer == "-" {
			return "", nil
		}
		if answer == "" {
			answer = ctx.Suggested
		}

		if strings.ContainsAny(answer, `/\`) {
			fmt.Println("A name cannot hold a path separator.")
			continue
		}
		if taken[answer] || config.Registered(answer) {
			fmt.Printf("Cluster %q already exists.\n", answer)
			continue
		}

		return answer, nil
	}
}

// ask prints a question and reads one line. promptui is not used here: its
// prompt either appends to the suggested name or loses it when the user presses
// Enter.
func ask(reader *bufio.Reader, question string) (string, error) {
	fmt.Print(question)

	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

// confirm asks a yes or no question. Anything but yes means no.
func confirm(reader *bufio.Reader, question string) bool {
	answer, err := ask(reader, question+" [y/N]: ")
	if err != nil {
		return false
	}

	answer = strings.ToLower(answer)

	return answer == "y" || answer == "yes"
}

// activeClusterName returns the cluster that holds the context the default
// kubectl config uses now. It returns an empty name when no imported cluster
// holds it.
func activeClusterName(registrations []registration) string {
	if importAs != "" {
		return importAs
	}

	active := config.ActiveContext(config.DefaultConfigFile())
	if active == "" {
		return ""
	}

	for _, r := range registrations {
		if r.context.Path == config.DefaultConfigFile() && r.context.Context == active {
			return r.name
		}
	}

	return ""
}

// takeOverDefaultConfig moves the default kubectl config aside and points it at
// the merged kubeconfig of a cluster.
func takeOverDefaultConfig(reader *bufio.Reader, cluster string) error {
	defaultConfigFile := config.DefaultConfigFile()
	backupFile := config.KubeBackupFile()

	if cluster == "" {
		fmt.Printf("%s is unchanged. Run `ks set-default <cluster-name>` to point it at ks.\n", defaultConfigFile)
		return nil
	}
	if config.Managed(defaultConfigFile) {
		fmt.Printf("%s already belongs to ks. It is unchanged.\n", defaultConfigFile)
		return nil
	}
	if _, err := os.Lstat(backupFile); err == nil {
		fmt.Printf("%s exists, so %s is unchanged.\n", backupFile, defaultConfigFile)
		fmt.Println("Move that backup away, or run `ks restore-kubeconfig` first.")
		return nil
	}

	_, err := os.Lstat(defaultConfigFile)
	hasDefault := err == nil

	if hasDefault && !importYes {
		fmt.Printf("\nks moves %s to %s.\n", defaultConfigFile, backupFile)
		fmt.Printf("It then points %s at the %q cluster.\n", defaultConfigFile, cluster)
		fmt.Println("`ks restore-kubeconfig` undoes this.")

		if !confirm(reader, "Move it") {
			fmt.Printf("%s is unchanged.\n", defaultConfigFile)
			return nil
		}
	}

	// Build the merged kubeconfig before moving anything, so the link never
	// points at a file that is not there.
	mergedFile, err := mergedConfigFile(cluster)
	if err != nil {
		return err
	}

	if hasDefault {
		if err := os.Rename(defaultConfigFile, backupFile); err != nil {
			return fmt.Errorf("failed to back up the default config: %w", err)
		}
	}

	if err := config.WriteDefaultRecord(cluster); err != nil {
		return err
	}
	if err := linkDefaultConfig(mergedFile); err != nil {
		return err
	}

	if hasDefault {
		fmt.Printf("Moved %s to %s.\n", defaultConfigFile, backupFile)
	}
	fmt.Printf("%s now points at the %q cluster.\n", defaultConfigFile, cluster)

	return nil
}
