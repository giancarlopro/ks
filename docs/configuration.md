# YAML Configuration Files

The `ks` CLI uses YAML configuration files to store information about Kubernetes clusters. These files are stored in the `~/.config/ks/clusters` folder, with each cluster having its own YAML configuration file named after the cluster.

## Structure of the YAML Configuration Files

Each YAML configuration file should follow the structure below:

```yaml
apiVersion: v1
clusters:
  - cluster:
      certificate-authority-data: <ca>
      server: <ip>
    name: <cluster-name>
contexts:
  - context:
      cluster: <cluster-name>
      namespace: <namespace>
      user: <user>
    name: <context-name>
current-context: <context-name>
kind: Config
preferences: {}
users:
  - name: <user>
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args: null
        command: gke-gcloud-auth-plugin
        env: null
        installHint: Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin
        interactiveMode: IfAvailable
        provideClusterInfo: true
```

### Fields

- `apiVersion`: The API version of the configuration. This field is required.
- `clusters`: A list of clusters, each containing the `certificate-authority-data` and `server` fields. This field is required.
- `contexts`: A list of contexts, each containing the `cluster`, `namespace`, and `user` fields. This field is required.
- `current-context`: The name of the current context. This field is required.
- `kind`: The kind of the configuration. This field is required.
- `preferences`: A map of preferences. This field is optional.
- `users`: A list of users, each containing the `exec` field with the `apiVersion`, `args`, `command`, `env`, `installHint`, `interactiveMode`, and `provideClusterInfo` fields. This field is required.

## Examples of Valid Configuration Files

### Example 1

```yaml
apiVersion: v1
clusters:
  - cluster:
      certificate-authority-data: abcdef1234567890
      server: https://my-cluster.example.com
    name: my-cluster
contexts:
  - context:
      cluster: my-cluster
      namespace: default
      user: my-cluster
    name: my-cluster
current-context: my-cluster
kind: Config
preferences: {}
users:
  - name: my-cluster
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args: null
        command: gke-gcloud-auth-plugin
        env: null
        installHint: Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin
        interactiveMode: IfAvailable
        provideClusterInfo: true
```

### Example 2

```yaml
apiVersion: v1
clusters:
  - cluster:
      certificate-authority-data: 0987654321fedcba
      server: https://another-cluster.example.com
    name: another-cluster
contexts:
  - context:
      cluster: another-cluster
      namespace: default
      user: another-cluster
    name: another-cluster
current-context: another-cluster
kind: Config
preferences: {}
users:
  - name: another-cluster
    user:
      exec:
        apiVersion: client.authentication.k8s.io/v1beta1
        args: null
        command: gke-gcloud-auth-plugin
        env: null
        installHint: Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin
        interactiveMode: IfAvailable
        provideClusterInfo: true
```

## Validation

The `ks` CLI validates the YAML configuration files when they are read or written. If a file is invalid or corrupted, a clear error message will be displayed, indicating the issue and the file path. Suggestions for fixing the issue will also be provided, such as checking the file format or using the `ks edit` command to correct the configuration.

## Merged kubeconfigs

The files in `~/.config/ks/clusters/` are the ones you write. `ks` also builds one **merged kubeconfig** for each of them, in `~/.config/ks/generated/`. A merged kubeconfig holds a context for every registered cluster. `ks activate` and `ks set-default` point at these files, not at the files you write.

Do not edit a merged kubeconfig by hand. Every rebuild overwrites it. Use `ks edit <cluster-name>` instead.

### Naming

Each context, cluster, and user in a merged kubeconfig takes the name of the file that registered it. A file named `prod.yaml` therefore produces a context named `prod`, whatever the provider called it.

A registered file with several contexts keeps all of them:

- The context named by `current-context` takes the registered name.
- Every other context takes `<cluster-name>/<original-context-name>`.
- When `current-context` is missing or names an unknown context, every context takes the second form.

Two contexts that share one cluster produce one cluster entry each, with the same content and different names.

### Rebuilds

`ks activate`, `ks <cluster-name>`, `ks set-default`, and `ks zsh-integration` rebuild every merged kubeconfig. A rebuild:

- reads every `.yaml` file in `~/.config/ks/clusters/`, so a file you add by hand needs no registration step;
- deletes any merged kubeconfig whose source file is gone;
- keeps the namespace of each context from the previous version of the same file, so a namespace set with `cns` or `kubectl config set-context` survives;
- skips a source file that does not parse, and prints `ks: skipping <path>: <reason>` on standard error.

`ks activate` fails only when the cluster you activate is itself unreadable, or holds no usable context.

### Fields

The merge copies every cluster, user, and context block through untouched. It rewrites only the names, the `cluster` and `user` references inside a context, and the `namespace`. Fields such as `client-certificate-data`, `token`, `insecure-skip-tls-verify`, and `proxy-url` therefore survive.

Merged kubeconfigs hold the same credentials as the files you write. The directory takes mode `0700`, and each file takes mode `0600`.

## Backup and Recovery

Before writing any changes to a YAML configuration file, the `ks` CLI creates a backup of the existing file. If an error occurs while reading or writing a file, the CLI will attempt to recover from the backup. The backup files are stored in the `~/.config/ks/clusters/backups` folder.

If the recovery process is unsuccessful, the user will be informed and provided with instructions on how to manually fix the issue if necessary.
