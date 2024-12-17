# ks CLI

The `ks` CLI is a command line tool for managing multiple Kubernetes configurations. It allows you to register, manage, and activate Kubernetes clusters, as well as integrate with Zsh and Oh-My-Zsh.

## Installation

Install the `ks` CLI using `go install`:
   ```sh
   go install github.com/giancarlopro/ks@latest
   ```

## Usage

The `ks` CLI provides the following commands:

- `ks`: Shows a list of clusters that you can select to enter an interactive shell with the correct environment variables.
- `ks add <cluster-name>`: Register a new Kubernetes cluster with the given name and open the default editor for the user to set the content of the file.
- `ks list`: List all registered Kubernetes clusters.
- `ks get <cluster-name>`: Get the details of a specific Kubernetes cluster.
- `ks edit <cluster-name>`: Open the default editor for the user to edit the cluster configuration.
- `ks delete <cluster-name>`: Delete a registered Kubernetes cluster.
- `ks activate <cluster-name>`: Activate a Kubernetes cluster, setting the `KUBECONFIG` environment variable to point to the selected cluster's configuration file.
- `ks set-default <cluster-name>`: Set a default Kubernetes cluster, creating a symbolic link to the default `kubectl` config file.
- `ks zsh-integration`: Set up Zsh and Oh-My-Zsh integration to read the `.ksconfig` file in a folder and set the cluster accordingly.

## Zsh and Oh-My-Zsh Integration

To set up Zsh and Oh-My-Zsh integration, follow these steps:

1. Create a `.ksconfig` file in the project directory to specify the desired Kubernetes cluster:
   ```sh
   echo "cluster-name" > .ksconfig
   ```

2. Add the following lines to your `.zshrc` file to enable the integration:
   ```sh
   ks zsh-integration
   ```

3. Restart your terminal or source the `.zshrc` file:
   ```sh
   source ~/.zshrc
   ```

Now, when you navigate to a directory with a `.ksconfig` file, the `KUBECONFIG` environment variable will be automatically set to the specified cluster's configuration file.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have any suggestions or improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
