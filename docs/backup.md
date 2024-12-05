# Backup and Recovery Process

The `ks` CLI provides a backup and recovery process to ensure that your Kubernetes cluster configurations are safe and can be restored in case of any issues.

## Backup Process

Before writing any changes to a YAML configuration file, the `ks` CLI creates a backup of the existing file. The backup files are stored in the `~/.config/ks/clusters/backups` folder.

### Steps to Create a Backup

1. When a cluster configuration is modified, the `ks` CLI reads the existing configuration file.
2. The CLI creates a backup directory (`~/.config/ks/clusters/backups`) if it doesn't already exist.
3. The existing configuration file is copied to the backup directory with a `.bak` extension.

## Recovery Process

If an error occurs while reading or writing a configuration file, the `ks` CLI will attempt to recover from the backup.

### Steps to Recover from a Backup

1. The CLI checks if a backup file exists in the `~/.config/ks/clusters/backups` folder.
2. If the backup file exists, the CLI reads the backup file.
3. The CLI writes the backup file content back to the original configuration file location.

## Manual Recovery

If the automatic recovery process is unsuccessful, you can manually recover the configuration file from the backup.

### Steps for Manual Recovery

1. Navigate to the `~/.config/ks/clusters/backups` folder.
2. Locate the backup file with the `.bak` extension corresponding to the cluster configuration you want to recover.
3. Copy the backup file to the `~/.config/ks/clusters` folder and rename it to remove the `.bak` extension.

### Example

```sh
cp ~/.config/ks/clusters/backups/my-cluster.yaml.bak ~/.config/ks/clusters/my-cluster.yaml
```

By following these steps, you can manually recover your Kubernetes cluster configuration from the backup file.
