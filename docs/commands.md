# ks CLI Commands

## ks

Shows a list of clusters that you can select to enter into an interactive shell with the correct environment variables.

### Usage

```sh
ks
```

## ks add

Register a new Kubernetes cluster with the given name and open the default editor so the user can set the content of the file.

### Usage

```sh
ks add <cluster-name>
```

### Example

```sh
ks add my-cluster
```

## ks list

List all registered Kubernetes clusters.

### Usage

```sh
ks list
```

### Example

```sh
ks list
```

## ks get

Get the details of a specific Kubernetes cluster.

### Usage

```sh
ks get <cluster-name>
```

### Example

```sh
ks get my-cluster
```

## ks edit

Open the default editor so the user can edit the cluster configuration.

### Usage

```sh
ks edit <cluster-name>
```

### Example

```sh
ks edit my-cluster
```

## ks delete

Delete a registered Kubernetes cluster.

### Usage

```sh
ks delete <cluster-name>
```

### Example

```sh
ks delete my-cluster
```

## ks activate

Activate a Kubernetes cluster, setting the `KUBECONFIG` environment variable to point to the selected cluster's configuration file.

### Usage

```sh
ks activate <cluster-name>
```

### Example

```sh
ks activate my-cluster
```

## ks set-default

Set a default Kubernetes cluster, creating a symbolic link to the default `kubectl` config file.

### Usage

```sh
ks set-default <cluster-name>
```

### Example

```sh
ks set-default my-cluster
```

## ks zsh-integration

Set up Zsh and Oh-My-Zsh integration to read the `.ksconfig` file in a folder and set the cluster accordingly.

### Usage

```sh
ks zsh-integration
```

### Example

```sh
ks zsh-integration
```
