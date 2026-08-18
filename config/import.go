package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// Import reads a kubeconfig that ks does not manage and turns it into cluster
// files. This file holds the part that reads and writes files. It prompts for
// nothing, so a test can drive it.

// SourceContext is one context of a source kubeconfig, ready to register.
type SourceContext struct {
	// Path is the source kubeconfig it comes from.
	Path string
	// Context is the name the source kubeconfig gives it.
	Context string
	// Suggested is the short name ks proposes to register it under.
	Suggested string
	// Server is the API server of its cluster, shown to the user.
	Server string

	// config is a kubeconfig that holds this context alone.
	config yaml.MapSlice
}

// Config returns a kubeconfig that holds this context, its cluster, and its
// user.
func (s SourceContext) Config() yaml.MapSlice {
	return s.config
}

// ksDir is the directory that holds everything ks manages.
func ksDir() string {
	return filepath.Join(HomeDir(), ".config", "ks")
}

// Managed reports whether a path lives inside the ks configuration directory.
// It follows links, so a link to a generated file counts as managed.
func Managed(path string) bool {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return false
	}

	dir, err := filepath.Abs(ksDir())
	if err != nil {
		return false
	}

	return resolved == dir || strings.HasPrefix(resolved, dir+string(filepath.Separator))
}

// SourcePaths returns the kubeconfigs to import. Arguments win. With no
// argument it reads KUBECONFIG, and falls back to the default kubectl
// configuration.
func SourcePaths(args []string) ([]string, error) {
	paths := args

	if len(paths) == 0 {
		if value := os.Getenv("KUBECONFIG"); value != "" {
			for _, path := range filepath.SplitList(value) {
				if strings.TrimSpace(path) != "" {
					paths = append(paths, path)
				}
			}
		}
	}
	if len(paths) == 0 {
		paths = []string{DefaultConfigFile()}
	}

	for _, path := range paths {
		if Managed(path) {
			return nil, fmt.Errorf("%s already belongs to ks", path)
		}
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("cannot read %s: %w", path, err)
		}
	}

	return paths, nil
}

// ReadSources turns every context of every source kubeconfig into a
// SourceContext. It returns a warning for each context it cannot use. It reads
// every file before it reports a parse error, so a caller writes nothing when
// one file is bad.
func ReadSources(paths []string) ([]SourceContext, []string, error) {
	var (
		contexts []SourceContext
		warnings []string
	)

	for _, path := range paths {
		kc, err := readKubeconfig(path)
		if err != nil {
			return nil, warnings, fmt.Errorf("cannot import %s: %w", path, err)
		}

		for _, ctx := range kc.contexts {
			body, ok := ctx.body.(yaml.MapSlice)
			if !ok {
				warnings = append(warnings, warning(path, fmt.Sprintf("context %q is not a mapping", ctx.name)))
				continue
			}

			clusterName := msString(body, "cluster")
			cluster, found := kc.clusters[clusterName]
			if !found {
				warnings = append(warnings, warning(path, fmt.Sprintf("context %q names unknown cluster %q", ctx.name, clusterName)))
				continue
			}

			userName := msString(body, "user")
			user, found := kc.users[userName]
			if !found {
				warnings = append(warnings, warning(path, fmt.Sprintf("context %q names unknown user %q", ctx.name, userName)))
				continue
			}

			contexts = append(contexts, SourceContext{
				Path:      path,
				Context:   ctx.name,
				Suggested: SuggestName(ctx.name),
				Server:    serverOf(cluster),
				config: kubeconfigDoc(
					[]interface{}{yaml.MapSlice{{Key: "cluster", Value: cluster}, {Key: "name", Value: clusterName}}},
					[]interface{}{yaml.MapSlice{{Key: "context", Value: body}, {Key: "name", Value: ctx.name}}},
					[]interface{}{yaml.MapSlice{{Key: "name", Value: userName}, {Key: "user", Value: user}}},
					ctx.name,
				),
			})
		}
	}

	return contexts, warnings, nil
}

// ReadSourcesAsOne returns a single kubeconfig that holds every context of
// every source kubeconfig. The first name wins when two files share one.
func ReadSourcesAsOne(paths []string) (yaml.MapSlice, error) {
	var (
		clusters, contexts, users []interface{}
		current                   string
	)
	seen := map[string]bool{}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot import %s: %w", path, err)
		}

		var doc yaml.MapSlice
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("cannot import %s: invalid YAML: %w", path, err)
		}

		for key, list := range map[string]*[]interface{}{
			"clusters": &clusters,
			"contexts": &contexts,
			"users":    &users,
		} {
			for _, item := range msList(doc, key) {
				entry, ok := item.(yaml.MapSlice)
				if !ok {
					continue
				}
				name := key + "/" + msString(entry, "name")
				if seen[name] {
					continue
				}
				seen[name] = true
				*list = append(*list, entry)
			}
		}

		if current == "" {
			current = msString(doc, "current-context")
		}
	}

	if len(contexts) == 0 {
		return nil, fmt.Errorf("no context to import")
	}

	return kubeconfigDoc(clusters, contexts, users, current), nil
}

// WriteCluster registers a cluster with the given kubeconfig.
func WriteCluster(name string, doc yaml.MapSlice) error {
	if err := os.MkdirAll(ClustersDir(), 0755); err != nil {
		return fmt.Errorf("failed to create clusters directory: %w", err)
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to render cluster %q: %w", name, err)
	}
	if err := os.WriteFile(ClusterConfigFile(name), data, 0600); err != nil {
		return fmt.Errorf("failed to write cluster %q: %w", name, err)
	}

	return nil
}

// Registered reports whether a cluster name is taken.
func Registered(name string) bool {
	_, err := os.Stat(ClusterConfigFile(name))
	return err == nil
}

// FreeName returns a name that no registered cluster uses. It appends a number
// to a name that is taken.
func FreeName(name string, taken map[string]bool) string {
	candidate := name
	for i := 2; Registered(candidate) || taken[candidate]; i++ {
		candidate = fmt.Sprintf("%s-%d", name, i)
	}
	return candidate
}

var unsafeName = regexp.MustCompile(`[^a-z0-9.-]+`)

// SuggestName derives a short cluster name from a context name. It takes the
// last segment, because every provider puts the cluster name there.
//
//	gke_acme_us-east1_prod                       -> prod
//	arn:aws:eks:us-east-1:123:cluster/staging    -> staging
func SuggestName(context string) string {
	segment := context
	if index := strings.LastIndexAny(segment, "_/:"); index >= 0 && index+1 < len(segment) {
		segment = segment[index+1:]
	}

	segment = unsafeName.ReplaceAllString(strings.ToLower(segment), "-")
	segment = strings.Trim(segment, "-.")

	if segment == "" {
		segment = unsafeName.ReplaceAllString(strings.ToLower(context), "-")
		segment = strings.Trim(segment, "-.")
	}
	if segment == "" {
		segment = "cluster"
	}

	return segment
}

// ActiveContext returns the current-context of a kubeconfig. It returns an
// empty string when the file is missing or unreadable.
func ActiveContext(path string) string {
	kc, err := readKubeconfig(path)
	if err != nil {
		return ""
	}
	return kc.currentContext
}

// kubeconfigDoc renders a kubeconfig from its three lists.
func kubeconfigDoc(clusters, contexts, users []interface{}, current string) yaml.MapSlice {
	return yaml.MapSlice{
		{Key: "apiVersion", Value: "v1"},
		{Key: "clusters", Value: clusters},
		{Key: "contexts", Value: contexts},
		{Key: "current-context", Value: current},
		{Key: "kind", Value: "Config"},
		{Key: "preferences", Value: yaml.MapSlice{}},
		{Key: "users", Value: users},
	}
}

// serverOf returns the API server of a cluster body.
func serverOf(cluster interface{}) string {
	body, ok := cluster.(yaml.MapSlice)
	if !ok {
		return ""
	}
	return msString(body, "server")
}
