package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// A kubeconfig is a Kubernetes client configuration file. It lists clusters,
// users, and contexts. A context names one cluster, one user, and a namespace.
//
// This file reads and writes kubeconfigs as generic YAML. The typed structs in
// config.go only model an exec user, so they drop fields such as
// client-certificate-data, token, and proxy-url. A merge rewrites the file, so
// it must keep every field it does not understand.

// namedNode is one entry of the clusters, users, or contexts list.
type namedNode struct {
	name string
	body interface{}
}

// kubeconfig holds the parts of a parsed kubeconfig that a merge needs.
type kubeconfig struct {
	currentContext string
	contexts       []namedNode
	clusters       map[string]interface{}
	users          map[string]interface{}
}

// MergedContext is one context of a merged kubeconfig, with the cluster body
// and the user body it points to.
type MergedContext struct {
	Name      string
	Namespace string

	cluster interface{}
	user    interface{}
	context yaml.MapSlice
}

// Contribution is what one registered cluster adds to every merged kubeconfig.
type Contribution struct {
	// Registered is the name of the source file, without the .yaml suffix.
	Registered string
	// Bare is the context that carries the registered name alone. It is empty
	// when no context earns that name.
	Bare string
	// Contexts are the contexts this source file adds, in sorted order.
	Contexts []MergedContext
	// Warnings describe the parts of the source file the merge skipped.
	Warnings []string
}

// readKubeconfig parses one kubeconfig file.
func readKubeconfig(path string) (*kubeconfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc yaml.MapSlice
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	kc := &kubeconfig{
		currentContext: msString(doc, "current-context"),
		clusters:       map[string]interface{}{},
		users:          map[string]interface{}{},
	}

	for _, item := range msList(doc, "clusters") {
		if name, body, ok := splitNamed(item, "cluster"); ok {
			kc.clusters[name] = body
		}
	}
	for _, item := range msList(doc, "users") {
		if name, body, ok := splitNamed(item, "user"); ok {
			kc.users[name] = body
		}
	}
	for _, item := range msList(doc, "contexts") {
		if name, body, ok := splitNamed(item, "context"); ok {
			kc.contexts = append(kc.contexts, namedNode{name: name, body: body})
		}
	}

	return kc, nil
}

// contribute turns one source file into the contexts it adds to every merged
// kubeconfig. It renames each context, cluster, and user after the registered
// name. A file with several contexts keeps the extra ones under
// "<registered>/<original context>".
func contribute(registered, path string, kc *kubeconfig) Contribution {
	c := Contribution{Registered: registered}

	// A context is usable only when its cluster and its user resolve inside
	// the same source file.
	var usable []namedNode
	for _, ctx := range kc.contexts {
		body, ok := ctx.body.(yaml.MapSlice)
		if !ok {
			c.Warnings = append(c.Warnings, warning(path, fmt.Sprintf("context %q is not a mapping", ctx.name)))
			continue
		}
		clusterRef := msString(body, "cluster")
		if _, found := kc.clusters[clusterRef]; !found {
			c.Warnings = append(c.Warnings, warning(path, fmt.Sprintf("context %q names unknown cluster %q", ctx.name, clusterRef)))
			continue
		}
		userRef := msString(body, "user")
		if _, found := kc.users[userRef]; !found {
			c.Warnings = append(c.Warnings, warning(path, fmt.Sprintf("context %q names unknown user %q", ctx.name, userRef)))
			continue
		}
		usable = append(usable, ctx)
	}

	if len(usable) == 0 {
		c.Warnings = append(c.Warnings, warning(path, "no usable context"))
		return c
	}

	// One context takes the registered name alone. With several contexts, that
	// is the current-context of the source file.
	bareOriginal := ""
	if len(usable) == 1 {
		bareOriginal = usable[0].name
	} else {
		for _, ctx := range usable {
			if ctx.name == kc.currentContext {
				bareOriginal = ctx.name
				break
			}
		}
	}

	for _, ctx := range usable {
		body := ctx.body.(yaml.MapSlice)

		name := registered + "/" + ctx.name
		if ctx.name == bareOriginal {
			name = registered
			c.Bare = name
		}

		rewritten := msSet(msSet(copyMapSlice(body), "cluster", name), "user", name)

		c.Contexts = append(c.Contexts, MergedContext{
			Name:      name,
			Namespace: msString(body, "namespace"),
			cluster:   kc.clusters[msString(body, "cluster")],
			user:      kc.users[msString(body, "user")],
			context:   rewritten,
		})
	}

	sort.Slice(c.Contexts, func(i, j int) bool {
		return c.Contexts[i].Name < c.Contexts[j].Name
	})

	return c
}

// buildMerged renders one merged kubeconfig. currentContext names the context
// the file activates. namespaces overrides the namespace of a context, so a
// namespace set with `kubectl config set-context` survives a rebuild.
func buildMerged(contexts []MergedContext, currentContext string, namespaces map[string]string) yaml.MapSlice {
	sorted := make([]MergedContext, len(contexts))
	copy(sorted, contexts)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	clusters := make([]interface{}, 0, len(sorted))
	contextList := make([]interface{}, 0, len(sorted))
	users := make([]interface{}, 0, len(sorted))

	for _, ctx := range sorted {
		clusters = append(clusters, yaml.MapSlice{
			{Key: "cluster", Value: ctx.cluster},
			{Key: "name", Value: ctx.Name},
		})

		body := copyMapSlice(ctx.context)
		if ns, found := namespaces[ctx.Name]; found && ns != "" {
			body = msSet(body, "namespace", ns)
		}
		contextList = append(contextList, yaml.MapSlice{
			{Key: "context", Value: body},
			{Key: "name", Value: ctx.Name},
		})

		users = append(users, yaml.MapSlice{
			{Key: "name", Value: ctx.Name},
			{Key: "user", Value: ctx.user},
		})
	}

	return yaml.MapSlice{
		{Key: "apiVersion", Value: "v1"},
		{Key: "clusters", Value: clusters},
		{Key: "contexts", Value: contextList},
		{Key: "current-context", Value: currentContext},
		{Key: "kind", Value: "Config"},
		{Key: "preferences", Value: yaml.MapSlice{}},
		{Key: "users", Value: users},
	}
}

// readNamespaces returns the namespace of each context in a kubeconfig file.
// It returns an empty map when the file is missing or unreadable.
func readNamespaces(path string) map[string]string {
	namespaces := map[string]string{}

	kc, err := readKubeconfig(path)
	if err != nil {
		return namespaces
	}

	for _, ctx := range kc.contexts {
		body, ok := ctx.body.(yaml.MapSlice)
		if !ok {
			continue
		}
		if ns := msString(body, "namespace"); ns != "" {
			namespaces[ctx.name] = ns
		}
	}

	return namespaces
}

func warning(path, reason string) string {
	return fmt.Sprintf("ks: skipping %s: %s", path, reason)
}

// splitNamed reads one entry of a clusters, users, or contexts list. It returns
// the entry name and the body stored under bodyKey.
func splitNamed(item interface{}, bodyKey string) (string, interface{}, bool) {
	entry, ok := item.(yaml.MapSlice)
	if !ok {
		return "", nil, false
	}
	name := msString(entry, "name")
	if strings.TrimSpace(name) == "" {
		return "", nil, false
	}
	body, found := msGet(entry, bodyKey)
	if !found {
		return "", nil, false
	}
	return name, body, true
}

// msGet returns the value stored under key.
func msGet(ms yaml.MapSlice, key string) (interface{}, bool) {
	for _, item := range ms {
		if k, ok := item.Key.(string); ok && k == key {
			return item.Value, true
		}
	}
	return nil, false
}

// msString returns the value stored under key as a string. It returns an empty
// string when the key is missing or holds another type.
func msString(ms yaml.MapSlice, key string) string {
	value, found := msGet(ms, key)
	if !found {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

// msList returns the value stored under key as a list.
func msList(ms yaml.MapSlice, key string) []interface{} {
	value, found := msGet(ms, key)
	if !found {
		return nil
	}
	list, ok := value.([]interface{})
	if !ok {
		return nil
	}
	return list
}

// msSet stores value under key. It keeps the position of an existing key and
// appends a new one.
func msSet(ms yaml.MapSlice, key string, value interface{}) yaml.MapSlice {
	for i, item := range ms {
		if k, ok := item.Key.(string); ok && k == key {
			ms[i].Value = value
			return ms
		}
	}
	return append(ms, yaml.MapItem{Key: key, Value: value})
}

// copyMapSlice copies the top level of a mapping, so a rewrite never changes
// the parsed source.
func copyMapSlice(ms yaml.MapSlice) yaml.MapSlice {
	out := make(yaml.MapSlice, len(ms))
	copy(out, ms)
	return out
}
