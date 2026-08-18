# Merged kubeconfig

## Problem Statement

`ks` keeps one kubeconfig file for each cluster. `ks activate` points `KUBECONFIG` at one of those files.

A kubeconfig file holds contexts. A context names a cluster, a user, and a namespace.

Many Kubernetes tools select a cluster by context name. `helmfile` uses `--kube-context`. A tool can only
select a context that the current kubeconfig holds.

With `ks`, each cluster lives in its own file. The current kubeconfig holds one context. So a tool cannot
reference any other cluster by name. Context names also come from the provider. A name such as
`gke_my-project_us-east1_prod` is long and does not match the name the user registered in `ks`.

## Solution

`ks` builds a merged kubeconfig for each registered cluster. A merged kubeconfig holds a context for every
registered cluster. Each context takes the name of the registered file.

`ks activate prod` points `KUBECONFIG` at the merged file for `prod`. That file sets `current-context` to
`prod`. The shell works on `prod` with no extra flags, as it does today.

The same file also holds `dev`, `staging`, and every other registered cluster. So `helmfile --kube-context
dev` works from that shell.

`ks set-default prod` links `~/.kube/config` to the merged file for `prod`. Plain `kubectl` gets the same
list of contexts.

## User Stories

1. As a user, I want each context to take the name I registered in `ks`, so that I write `--kube-context prod`
   and not a long provider name.
2. As a user, I want the merged file to hold every registered cluster, so that one shell can reach another
   cluster by context name.
3. As a user, I want the activated cluster to be the `current-context`, so that plain `kubectl` still works
   on one cluster for each shell.
4. As a user, I want `ks <name>` and `ks activate <name>` to build the merged file, so that I never run a
   separate sync command.
5. As a user, I want `ks set-default <name>` to link `~/.kube/config` to the merged file, so that tools
   outside a `ks` shell see the same contexts.
6. As a user, I want a registered file with several contexts to keep all of them, so that no cluster
   disappears from the merged file.
7. As a user, I want a file I drop into the clusters directory by hand to appear on the next `ks` run, so
   that I do not run a registration step.
8. As a user, I want a broken registered file to be skipped with a warning, so that one bad file does not
   block every shell.
9. As a user, I want a clear error when the cluster I activate is the broken one, so that I know why the
   shell did not start.
10. As a user, I want the namespace I set with `cns` to survive the next activation, so that the tool keeps
    the behaviour it has today.
11. As a user, I want a deleted cluster to leave no merged file behind, so that stale contexts do not appear.
12. As a user, I want fields such as `client-certificate-data`, `token`, and `proxy-url` to survive the
    merge, so that every authentication method keeps working.

## Implementation Decisions

### Vocabulary

- **source file**: a registered kubeconfig in the clusters directory, named `<name>.yaml`.
- **registered name**: the file name of a source file, without the `.yaml` suffix.
- **generated file**: a merged kubeconfig that `ks` writes, named `<name>.yaml`.
- **rebuild**: the step that writes one generated file for every source file.

### Paths

- Source files stay in `~/.config/ks/clusters/`.
- Generated files go in `~/.config/ks/generated/`.
- The generated directory takes mode `0700`. Generated files take mode `0600`.
- The paths module gains functions for the generated directory and for a generated file path.

### Merge fidelity

The merge uses generic YAML maps. It does not use the typed structs in the config module. Those structs
only model an `exec` user and drop every other field.

The merge copies each cluster block, user block, and context block through untouched. It rewrites only the
`name` fields, the `cluster` and `user` reference strings inside a context, and the `namespace` field.

The build adds no new dependency. It uses the YAML library the project already has.

### Naming rule

For each source file, in order of registered name:

- A source file with one context contributes that context under the registered name.
- A source file with several contexts contributes its `current-context` under the registered name. It
  contributes every other context under `<registered name>/<original context name>`.
- A source file with several contexts and no valid `current-context` contributes every context under
  `<registered name>/<original context name>`.
- A source file with no contexts contributes nothing. `ks` prints a warning.

For every contributed context, the merged file gains one cluster entry and one user entry. Both take the
same name as the context. A source file that shares one cluster between two contexts therefore produces two
cluster entries with the same content and different names.

A context whose `cluster` or `user` reference does not resolve inside its own source file is skipped with a
warning.

Contexts appear in the merged file sorted by name. Cluster and user entries follow the same order. The
output is therefore the same on every rebuild.

### Merged file shape

- `apiVersion` is `v1`.
- `kind` is `Config`.
- `preferences` is an empty map.
- `current-context` is the registered name of the cluster the file belongs to.
- When that cluster contributes no bare name, `current-context` is the first context of that source file in
  sort order.

### Rebuild scope

`ks activate <name>`, `ks <name>`, `ks set-default <name>`, and `ks zsh-integration` all rebuild every
generated file.

A rebuild also deletes any generated file that has no source file. It leaves other files in the directory
alone.

A rebuild writes each file to a temporary file in the generated directory, then renames it. A concurrent
`kubectl` therefore never reads a half-written file.

### Namespace carry-forward

Before it writes a generated file, the rebuild reads the current version of that same generated file.

For each context that exists in both versions, the new version keeps the namespace from the old version.
A context that is new takes the namespace from its source file.

An unreadable old generated file is ignored. The rebuild then uses source namespaces.

This keeps the behaviour of the `cns` shell function, which writes a namespace into the file that
`KUBECONFIG` names.

### Failure handling

A source file that fails to parse is skipped. `ks` prints one line to standard error:
`ks: skipping <path>: <reason>`.

The command still fails when the cluster the user activates is itself unreadable, or contributes no context.

### Commands

- `ks activate <name>` and `ks <name>` rebuild, then spawn the shell with `KUBECONFIG` set to the generated
  file for `<name>`.
- `ks set-default <name>` rebuilds, then links `~/.kube/config` to the generated file for `<name>`. The
  Windows fallback copies that file, as it does today.
- `ks set-default <name>` also records the name in `~/.config/ks/default`. A rebuild reads that record. On
  Windows it refreshes the copy at `~/.kube/config` from the matching generated file. On other systems the
  symbolic link needs no refresh.
- `ks zsh-integration` rebuilds, then points `KUBECONFIG` at the generated file. It also trims the name it
  reads from `.ksconfig`, because `echo` writes a newline.
- `ks add`, `ks get`, `ks edit`, and `ks delete` keep reading and writing source files only.

### Listing rules

The rebuild reads the same set of files that `ks list` reads. It skips directories, so the `backups`
directory stays excluded. It reads only files with a `.yaml` suffix.

## Testing Decisions

The seam is the `HOME` environment variable. The paths module derives every directory from it. A test sets
`HOME` to a temporary directory, writes source files, and calls the rebuild. No other part is replaced.
The set-default test already uses this seam.

Tests live in the config package, beside the merge code. They read the generated YAML back with the YAML
library and assert on the parsed maps. They do not compare raw strings.

Cover this behaviour:

- A single-context source file produces a context, a cluster, and a user under the registered name.
- The merged file for one cluster holds the contexts of every other cluster.
- `current-context` is the registered name of the cluster the file belongs to.
- A source file with three contexts produces one bare name and two `<name>/<context>` names.
- A source file with no valid `current-context` produces only `<name>/<context>` names.
- A field the typed structs do not model, such as `token`, survives the merge.
- A broken source file is skipped, and the other clusters still appear.
- A rebuild after a namespace change in a generated file keeps the new namespace.
- A rebuild deletes the generated file of a source file that no longer exists.
- Two contexts that share one cluster produce two cluster entries.

Command tests stay light. They check that activate points `KUBECONFIG` at the generated path, and that
set-default links `~/.kube/config` to the generated path.

### Integration tests

The unit tests read the generated file with a YAML library. That proves the file says what `ks` means. It
does not prove that a Kubernetes tool accepts it.

A second set of tests loads the generated file with `kubectl`. They carry the `integration` build tag, so
the existing CI step for that tag runs them. They need no cluster, because they only read the file.

`kubectl` is the right tool to test with. `helmfile` does not read kubeconfigs. It passes `--kube-context`
to `helm`, and `helm` gives the name to the client-go loader. `kubectl` uses that same loader. So a context
that `kubectl` resolves is a context `helmfile` can select, and testing through `helmfile` would add two
binaries and a fixture to reach the same loader.

Cover this behaviour:

- `kubectl config get-contexts` lists exactly the registered names.
- `kubectl config current-context` names the activated cluster.
- `kubectl config view --minify --context=<name>` resolves each context, with the right server and user.
  This fails when the rename leaves a cluster or a user reference behind.
- `kubectl config view --minify` fails for an original provider context name, which proves the rename is
  complete and proves the checks above can fail.

The tests skip when `kubectl` is missing. A `KS_KUBECTL` environment variable names a `kubectl` that is not
on the path.

## Out of Scope

- A command that only rebuilds, such as `ks sync`.
- Editing a merged file. Source files stay the only input.
- Merging kubeconfigs from outside the clusters directory.
- Changing the `ks add` template.
- Changing the `cns` function or any other shell alias.
- Removing the typed structs in the config module. Other commands still use them.

## Further Notes

The generated file holds the same credentials as the source files. It gains mode `0600`, which matches the
mode that `ks add` writes.

Two shells on the same cluster share one generated file. This matches the behaviour today, where they share
one source file.

The old behaviour pointed `KUBECONFIG` at a source file. A user with a script that reads that path must
change it to the generated path. The README and the command documentation must state the new path.
