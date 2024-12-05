# YAML Configuration Files

The `ks` CLI uses YAML configuration files to store information about Kubernetes clusters. These files are stored in the `~/.config/ks/clusters` folder, with each cluster having its own YAML configuration file named after the cluster.

## Structure of the YAML Configuration Files

Each YAML configuration file should follow the structure below:

```yaml
name: <cluster-name>
server: <server-url>
token: <authentication-token>
```

### Fields

- `name`: The name of the Kubernetes cluster. This field is required.
- `server`: The URL of the Kubernetes API server. This field is required.
- `token`: The authentication token for accessing the Kubernetes API server. This field is required.

## Examples of Valid Configuration Files

### Example 1

```yaml
name: my-cluster
server: https://my-cluster.example.com
token: abcdef1234567890
```

### Example 2

```yaml
name: another-cluster
server: https://another-cluster.example.com
token: 0987654321fedcba
```

## Validation

The `ks` CLI validates the YAML configuration files when they are read or written. If a file is invalid or corrupted, a clear error message will be displayed, indicating the issue and the file path. Suggestions for fixing the issue will also be provided, such as checking the file format or using the `ks edit` command to correct the configuration.

## Backup and Recovery

Before writing any changes to a YAML configuration file, the `ks` CLI creates a backup of the existing file. If an error occurs while reading or writing a file, the CLI will attempt to recover from the backup. The backup files are stored in the `~/.config/ks/clusters/backups` folder.

If the recovery process is unsuccessful, the user will be informed and provided with instructions on how to manually fix the issue if necessary.
