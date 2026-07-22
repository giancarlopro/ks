package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// aliasDef is a short kubectl shortcut: `name` expands to `command` (plus any
// extra arguments the user passes).
type aliasDef struct {
	name    string
	command string
}

// kubectlAliases is the curated set of kubectl shortcuts ks knows how to install.
var kubectlAliases = []aliasDef{
	{"kg", "kubectl get"},
	{"kd", "kubectl describe"},
	{"krrd", "kubectl rollout restart deployment"},
	{"krrs", "kubectl rollout restart statefulset"},
	{"ktn", "kubectl top node"},
	{"kdn", "kubectl describe node"},
	{"kl", "kubectl logs"},
	{"klf", "kubectl logs -f"},
}

var shellInitCmd = &cobra.Command{
	Use:   "shell-init [shell]",
	Short: "Print kubectl alias/function definitions for your shell",
	Long: `Print a set of handy kubectl aliases and functions for the given shell.

Supported shells: bash, zsh, powershell, cmd. If no shell is given, ks tries
to detect it from the environment.

Aliases: kg (get), kd (describe), krrd/krrs (rollout restart deployment/
statefulset), ktn (top node), kdn (describe node), kl (logs), klf (logs -f).
Plus a cns function to switch the current namespace (with an fzf picker when
no namespace is given).

Add the output to your shell startup so it loads every session:

  # bash (~/.bashrc)
  eval "$(ks shell-init bash)"

  # zsh (~/.zshrc)
  eval "$(ks shell-init zsh)"

  # PowerShell ($PROFILE)
  ks shell-init powershell | Out-String | Invoke-Expression

  # cmd.exe (persist via AutoRun)
  ks shell-init cmd > "%USERPROFILE%\ks-aliases.cmd"
  reg add "HKCU\Software\Microsoft\Command Processor" /v AutoRun /d "%USERPROFILE%\ks-aliases.cmd" /f`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := detectShell()
		if len(args) == 1 {
			shell = args[0]
		}

		script, err := aliasScript(shell)
		if err != nil {
			return err
		}

		fmt.Print(script)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}

// detectShell guesses the current shell from the environment.
func detectShell() string {
	if runtime.GOOS == "windows" {
		// PSModulePath is set inside PowerShell; otherwise assume cmd.exe.
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		return "cmd"
	}

	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return "zsh"
	case strings.Contains(shell, "bash"):
		return "bash"
	default:
		return "bash"
	}
}

// aliasScript returns the alias/function definitions for the given shell name.
func aliasScript(shell string) (string, error) {
	switch strings.ToLower(shell) {
	case "bash", "zsh", "sh":
		return posixAliases(), nil
	case "powershell", "pwsh":
		return powershellAliases(), nil
	case "cmd", "cmd.exe":
		return cmdAliases(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, powershell, cmd)", shell)
	}
}

// posixAliases renders the shortcuts for bash and zsh.
func posixAliases() string {
	var b strings.Builder
	b.WriteString("# ks kubectl aliases\n")
	for _, a := range kubectlAliases {
		fmt.Fprintf(&b, "alias %s=%q\n", a.name, a.command)
	}
	b.WriteString(`
cns() {
  local namespace
  if [ "$#" -eq 1 ]; then
    namespace="$1"
  else
    namespace="$(kubectl get namespace -o custom-columns=":metadata.name" | fzf)"
  fi
  if [ -z "$namespace" ]; then
    return 0
  fi
  kubectl config set-context --current --namespace "$namespace"
}
`)
	return b.String()
}

// powershellAliases renders the shortcuts as PowerShell functions. PowerShell
// aliases cannot carry arguments, so each shortcut becomes a function.
func powershellAliases() string {
	var b strings.Builder
	b.WriteString("# ks kubectl aliases\n")
	for _, a := range kubectlAliases {
		fmt.Fprintf(&b, "function %s { %s @args }\n", a.name, a.command)
	}
	b.WriteString(`
function cns {
  param([string]$namespace)
  if (-not $namespace) {
    $namespace = kubectl get namespace -o custom-columns=":metadata.name" | fzf
  }
  if (-not $namespace) { return }
  kubectl config set-context --current --namespace $namespace
}
`)
	return b.String()
}

// cmdAliases renders the shortcuts as cmd.exe doskey macros. doskey cannot
// express the fzf-driven picker, so cns requires an explicit namespace here.
func cmdAliases() string {
	var b strings.Builder
	b.WriteString("@echo off\n")
	for _, a := range kubectlAliases {
		fmt.Fprintf(&b, "doskey %s=%s $*\n", a.name, a.command)
	}
	b.WriteString("doskey cns=kubectl config set-context --current --namespace $1\n")
	return b.String()
}
