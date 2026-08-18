//go:build integration

package config

import (
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

// These tests load a merged kubeconfig with kubectl. kubectl uses the same
// loader as helm, and helmfile passes --kube-context to helm. So a context that
// kubectl resolves is a context helmfile can select.
//
// The tests need no cluster. They only read the file.
//
// Run them with: go test -tags=integration ./...
// Set KS_KUBECTL to use a kubectl that is not on PATH.

// prodSource is a kubeconfig as a cloud provider writes it. The names are long
// and do not match the name the user registers.
const prodSource = `apiVersion: v1
kind: Config
preferences: {}
clusters:
- cluster:
    certificate-authority-data: ZmFrZS1jYQ==
    server: https://prod.example.com
  name: gke_project_region_prod
contexts:
- context:
    cluster: gke_project_region_prod
    namespace: default
    user: gke_project_region_prod
  name: gke_project_region_prod
current-context: gke_project_region_prod
users:
- name: gke_project_region_prod
  user:
    token: prod-token
`

// requireKubectl returns the kubectl to test with, or skips the test.
func requireKubectl(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("KS_KUBECTL"); path != "" {
		return path
	}
	path, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl is not installed")
	}
	return path
}

// kubectlRun runs kubectl against one kubeconfig and returns its output.
func kubectlRun(t *testing.T, kubectl, kubeconfig string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(kubectl, args...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)

	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// kubectlOut runs kubectl and fails the test when kubectl reports an error.
func kubectlOut(t *testing.T, kubectl, kubeconfig string, args ...string) string {
	t.Helper()

	out, err := kubectlRun(t, kubectl, kubeconfig, args...)
	if err != nil {
		t.Fatalf("kubectl %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return out
}

func TestKubectlReadsMergedConfig(t *testing.T) {
	kubectl := requireKubectl(t)

	useTempHome(t)
	writeSource(t, "prod", prodSource)
	writeSource(t, "lab", threeContexts)
	rebuild(t)

	merged := GeneratedConfigFile("prod")

	// Every registered cluster is addressable by the name ks knows it under.
	// This is what a tool needs for --kube-context.
	names := strings.Split(kubectlOut(t, kubectl, merged, "config", "get-contexts", "-o", "name"), "\n")
	sort.Strings(names)
	want := "lab,lab/alpha,lab/gamma,prod"
	if got := strings.Join(names, ","); got != want {
		t.Errorf("expected contexts %q, got %q", want, got)
	}

	// The activated cluster is the one a command uses without a flag.
	if got := kubectlOut(t, kubectl, merged, "config", "current-context"); got != "prod" {
		t.Errorf("expected current-context prod, got %q", got)
	}
}

func TestKubectlResolvesEveryContext(t *testing.T) {
	kubectl := requireKubectl(t)

	useTempHome(t)
	writeSource(t, "prod", prodSource)
	writeSource(t, "lab", threeContexts)
	rebuild(t)

	merged := GeneratedConfigFile("prod")

	// A context resolves only when its cluster and its user resolve with it.
	// The rename must therefore line up all three names.
	servers := map[string]string{
		"prod":      "https://prod.example.com",
		"lab":       "https://two.example.com",
		"lab/alpha": "https://one.example.com",
		"lab/gamma": "https://one.example.com",
	}

	for name, server := range servers {
		got := kubectlOut(t, kubectl, merged, "config", "view", "--minify",
			"--context="+name, "-o", "jsonpath={.clusters[0].cluster.server}")
		if got != server {
			t.Errorf("context %q: expected server %q, got %q", name, server, got)
		}

		got = kubectlOut(t, kubectl, merged, "config", "view", "--minify",
			"--context="+name, "-o", "jsonpath={.users[0].name}")
		if got != name {
			t.Errorf("context %q: expected user %q, got %q", name, name, got)
		}
	}
}

func TestKubectlDoesNotSeeOriginalContextNames(t *testing.T) {
	kubectl := requireKubectl(t)

	useTempHome(t)
	writeSource(t, "prod", prodSource)
	rebuild(t)

	merged := GeneratedConfigFile("prod")

	// The provider name is gone, so the rename is complete. This also proves
	// that the checks above fail when a context is missing.
	out, err := kubectlRun(t, kubectl, merged, "config", "view", "--minify",
		"--context=gke_project_region_prod")
	if err == nil {
		t.Errorf("expected the original context name to be gone, got:\n%s", out)
	}
}

func TestKubectlReadsNamespaceOfActivatedContext(t *testing.T) {
	kubectl := requireKubectl(t)

	useTempHome(t)
	writeSource(t, "prod", prodSource)
	rebuild(t)

	merged := GeneratedConfigFile("prod")

	got := kubectlOut(t, kubectl, merged, "config", "view", "--minify",
		"-o", "jsonpath={.contexts[0].context.namespace}")
	if got != "default" {
		t.Errorf("expected namespace default, got %q", got)
	}
}
