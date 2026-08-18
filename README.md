# ks CLI

The `ks` CLI is a command line tool for managing multiple Kubernetes configurations. It allows you to register, manage, and activate Kubernetes clusters, as well as integrate with Zsh and Oh-My-Zsh.

It runs on Linux, macOS, and Windows. On Windows it uses `%USERPROFILE%` for the home directory, spawns the shell defined by `ComSpec` (defaulting to `cmd.exe`), defaults to `notepad` as the editor, and copies the default kubeconfig when symlinks aren't permitted.

## Installation

Install the `ks` CLI using `go install`:

```sh
go install github.com/giancarlopro/ks@latest
```

## Usage

The `ks` CLI provides the following commands:

- `ks`: Shows a list of clusters that you can select to enter an interactive shell with the correct environment variables.
- `ks <cluster-name>`: Activate a cluster directly, without opening the selector, entering an interactive shell with the correct environment variables.
- `ks add <cluster-name>`: Register a new Kubernetes cluster with the given name and open the default editor for the user to set the content of the file.
- `ks list`: List all registered Kubernetes clusters.
- `ks get <cluster-name>`: Get the details of a specific Kubernetes cluster.
- `ks edit <cluster-name>`: Open the default editor for the user to edit the cluster configuration.
- `ks delete <cluster-name>`: Delete a registered Kubernetes cluster.
- `ks activate <cluster-name>`: Activate a Kubernetes cluster, setting the `KUBECONFIG` environment variable to point to the merged kubeconfig for that cluster.
- `ks set-default <cluster-name>`: Set a default Kubernetes cluster, creating a symbolic link from the default `kubectl` config file to the merged kubeconfig for that cluster.
- `ks zsh-integration`: Set up Zsh and Oh-My-Zsh integration to read the `.ksconfig` file in a folder and set the cluster accordingly.
- `ks shell-init [shell]`: Print a set of handy `kubectl` aliases/functions for the given shell (`bash`, `zsh`, `powershell`, or `cmd`). The shell is auto-detected when omitted.

## Merged kubeconfig

Each cluster you register keeps its own file in `~/.config/ks/clusters/`. Many tools, such as `helmfile`,
select a cluster by context name instead of by file. A tool can only reach a context that the current
kubeconfig holds, so one file for each cluster leaves those tools with one context.

`ks` therefore builds a **merged kubeconfig** for every registered cluster, in `~/.config/ks/generated/`.
A merged kubeconfig holds a context for every registered cluster, and each context takes the name you
registered. `ks activate prod` points `KUBECONFIG` at `~/.config/ks/generated/prod.yaml`, which sets
`current-context: prod`.

Your shell still works on one cluster, as before:

```sh
ks prod
kubectl get pods          # runs against prod
helmfile --kube-context dev apply   # reaches another cluster by name
```

The merged files are rebuilt on every `ks activate`, `ks <cluster-name>`, `ks set-default`, and
`ks zsh-integration`. So a file you drop into `~/.config/ks/clusters/` by hand appears on the next run,
and a cluster you delete leaves no merged file behind.

Details worth knowing:

- A registered file with several contexts keeps all of them. The `current-context` of that file takes the
  registered name. Every other context takes `<cluster-name>/<original-context-name>`.
- A registered file that does not parse is skipped with a warning on standard error. The other clusters
  still work. `ks activate` fails only when the cluster you activate is itself the broken one.
- A namespace you set with `cns`, or with `kubectl config set-context`, survives the next rebuild.
- Never edit a file in `~/.config/ks/generated/`. Use `ks edit <cluster-name>`, which edits the source file.

## kubectl Aliases

`ks shell-init` emits a curated set of `kubectl` shortcuts for your shell:

| Alias | Expands to |
| --- | --- |
| `kg` | `kubectl get` |
| `kd` | `kubectl describe` |
| `krrd` | `kubectl rollout restart deployment` |
| `krrs` | `kubectl rollout restart statefulset` |
| `ktn` | `kubectl top node` |
| `kdn` | `kubectl describe node` |
| `kl` | `kubectl logs` |
| `klf` | `kubectl logs -f` |
| `cns [namespace]` | Switch the current namespace (falls back to an `fzf` picker when no namespace is given) |

To load them every session, add the appropriate line to your shell startup file:

```sh
# bash (~/.bashrc)
eval "$(ks shell-init bash)"

# zsh (~/.zshrc)
eval "$(ks shell-init zsh)"
```

```powershell
# PowerShell ($PROFILE)
ks shell-init powershell | Out-String | Invoke-Expression
```

```bat
:: cmd.exe — write the doskey macros and load them on startup via AutoRun
ks shell-init cmd > "%USERPROFILE%\ks-aliases.cmd"
reg add "HKCU\Software\Microsoft\Command Processor" /v AutoRun /d "%USERPROFILE%\ks-aliases.cmd" /f
```

> On `cmd.exe`, `cns` requires an explicit namespace argument, since `doskey` cannot run the `fzf` picker.

## Zsh and Oh-My-Zsh Integration

To set up Zsh and Oh-My-Zsh integration, follow these steps:

1. Create a `.ksconfig` file in the project directory to specify the desired Kubernetes cluster:

```sh
echo "cluster-name" > .ksconfig
```

1. Add the following lines to your `.zshrc` file to enable the integration:

```sh
ks zsh-integration
```

1. Restart your terminal or source the `.zshrc` file:

```sh
source ~/.zshrc
```

Now, when you navigate to a directory with a `.ksconfig` file, the `KUBECONFIG` environment variable will be automatically set to the merged kubeconfig for the specified cluster.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have any suggestions or improvements.

Run the unit tests with `go test ./...`. The integration tests load a generated kubeconfig with `kubectl`, to prove that tools such as `helmfile` can select its contexts. They need no cluster:

```sh
go test -tags=integration ./...
```

They skip when `kubectl` is missing. Set `KS_KUBECTL` to use a `kubectl` that is not on your `PATH`.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
