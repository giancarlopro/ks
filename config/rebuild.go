package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// RebuildResult reports what a rebuild produced.
type RebuildResult struct {
	// Built maps a registered cluster name to its merged kubeconfig path.
	Built map[string]string
	// Skipped maps a registered cluster name to the reason it has no merged
	// kubeconfig.
	Skipped map[string]error
}

// Path returns the merged kubeconfig of a cluster, and the reason the rebuild
// skipped it.
func (r *RebuildResult) Path(name string) (string, error) {
	if path, found := r.Built[name]; found {
		return path, nil
	}
	if err, found := r.Skipped[name]; found {
		return "", err
	}
	return "", fmt.Errorf("cluster %q does not exist", name)
}

// Rebuild writes one merged kubeconfig for every registered cluster. Each
// merged kubeconfig holds the contexts of every registered cluster, renamed
// after the file that registered them. It sets current-context to the cluster
// the file belongs to.
//
// Rebuild writes warnings about unusable source files to warnOut. A nil
// warnOut discards them.
func Rebuild(warnOut io.Writer) (*RebuildResult, error) {
	if warnOut == nil {
		warnOut = io.Discard
	}

	if err := os.MkdirAll(ClustersDir(), 0755); err != nil {
		return nil, fmt.Errorf("failed to create clusters directory: %w", err)
	}
	if err := os.MkdirAll(GeneratedDir(), 0700); err != nil {
		return nil, fmt.Errorf("failed to create generated directory: %w", err)
	}

	names, err := registeredNames()
	if err != nil {
		return nil, err
	}

	result := &RebuildResult{
		Built:   map[string]string{},
		Skipped: map[string]error{},
	}

	var (
		contributions []Contribution
		all           []MergedContext
	)

	for _, name := range names {
		path := ClusterConfigFile(name)

		kc, err := readKubeconfig(path)
		if err != nil {
			fmt.Fprintln(warnOut, warning(path, err.Error()))
			result.Skipped[name] = fmt.Errorf("cluster %q is unreadable: %w", name, err)
			continue
		}

		contribution := contribute(name, path, kc)
		for _, w := range contribution.Warnings {
			fmt.Fprintln(warnOut, w)
		}
		if len(contribution.Contexts) == 0 {
			result.Skipped[name] = fmt.Errorf("cluster %q holds no usable context", name)
			continue
		}

		contributions = append(contributions, contribution)
		all = append(all, contribution.Contexts...)
	}

	for _, contribution := range contributions {
		current := contribution.Bare
		if current == "" {
			current = contribution.Contexts[0].Name
		}

		path := GeneratedConfigFile(contribution.Registered)
		doc := buildMerged(all, current, readNamespaces(path))

		data, err := yaml.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to render merged config for %q: %w", contribution.Registered, err)
		}
		if err := writeAtomic(path, data); err != nil {
			return nil, err
		}

		result.Built[contribution.Registered] = path
	}

	if err := pruneGenerated(names); err != nil {
		return nil, err
	}

	if err := refreshDefaultCopy(result); err != nil {
		return nil, err
	}

	return result, nil
}

// registeredNames lists the registered clusters, sorted by name.
func registeredNames() ([]string, error) {
	entries, err := os.ReadDir(ClustersDir())
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		names = append(names, strings.TrimSuffix(entry.Name(), ".yaml"))
	}
	sort.Strings(names)

	return names, nil
}

// pruneGenerated removes merged kubeconfigs whose source file is gone.
func pruneGenerated(names []string) error {
	registered := map[string]bool{}
	for _, name := range names {
		registered[name] = true
	}

	entries, err := os.ReadDir(GeneratedDir())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		if registered[strings.TrimSuffix(entry.Name(), ".yaml")] {
			continue
		}
		if err := os.Remove(filepath.Join(GeneratedDir(), entry.Name())); err != nil {
			return fmt.Errorf("failed to remove stale merged config: %w", err)
		}
	}

	return nil
}

// refreshDefaultCopy updates the default kubectl config on systems that copy it
// instead of linking it. A symbolic link needs no refresh.
func refreshDefaultCopy(result *RebuildResult) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	data, err := os.ReadFile(DefaultRecordFile())
	if err != nil {
		return nil
	}
	name := strings.TrimSpace(string(data))

	path, found := result.Built[name]
	if !found {
		return nil
	}

	info, err := os.Lstat(DefaultConfigFile())
	if err != nil || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	return CopyFile(path, DefaultConfigFile())
}

// WriteDefaultRecord records the cluster that `ks set-default` selected.
func WriteDefaultRecord(name string) error {
	if err := os.MkdirAll(filepath.Dir(DefaultRecordFile()), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if err := os.WriteFile(DefaultRecordFile(), []byte(name+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to record default cluster: %w", err)
	}
	return nil
}

// writeAtomic writes data to a temporary file, then renames it. A reader
// therefore never sees a half-written file.
func writeAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".ks-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	name := tmp.Name()

	defer os.Remove(name)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("failed to write merged config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to write merged config: %w", err)
	}
	if err := os.Chmod(name, 0600); err != nil {
		return fmt.Errorf("failed to set merged config permissions: %w", err)
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("failed to replace merged config: %w", err)
	}

	return nil
}

// CopyFile copies the contents of src to dst, creating or truncating dst.
func CopyFile(src, dst string) error {
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
