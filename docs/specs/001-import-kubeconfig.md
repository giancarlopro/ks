# Import a kubeconfig

## Problem Statement

A new `ks` user already has a kubeconfig. It holds every cluster they work with, in one file that `ks` does
not manage.

`ks` registers one cluster for each file in its clusters directory. To start using `ks`, the user must
split that kubeconfig by hand. They must copy each context, with the cluster and the user it names, into a
file of its own. This is slow and easy to get wrong.

The provider also names contexts. A name such as `gke_acme_us-east1_prod` is long. A user who copies it by
hand keeps that name, which is what `ks` short names exist to avoid.

## Solution

`ks import` reads a kubeconfig that `ks` does not manage. It finds every context. It asks the user for a
short name for each one, and suggests a name it derives from the context name. It writes one cluster file
for each context.

Import then takes over the default kubectl configuration. It moves `~/.kube/config` to
`~/.kube/config.bkp`. It links `~/.kube/config` to the merged kubeconfig of the cluster that held the
active context, so `kubectl` keeps working on the same cluster.

Import states this move before it makes it, and waits for the user to confirm.

`ks restore-kubeconfig` puts the original file back. The clusters that import registered stay registered.

## User Stories

1. As a new user, I want one command to register every cluster in my kubeconfig, so that I do not split the
   file by hand.
2. As a new user, I want to give each context a short name, so that I type `ks prod` and not the name my
   cloud provider chose.
3. As a new user, I want a suggested name for each context, so that I press Enter and move on.
4. As a new user, I want to skip a context, so that I do not register a cluster I no longer use.
5. As a new user, I want `kubectl` to keep working on the same cluster after the import, so that nothing in
   my shell breaks.
6. As a new user, I want to be told before my kubeconfig moves, so that the change is never a surprise.
7. As a new user, I want one command to undo the move, so that I can go back to my old setup.
8. As a user with `KUBECONFIG` set to several files, I want import to read all of them, so that I register
   every cluster in one run.
9. As a user, I want import to refuse a file that `ks` already manages, so that I never import my own
   generated files.
10. As a user, I want import to keep my existing registered clusters, so that a second import never
    overwrites my work.
11. As a scripting user, I want a flag that accepts every suggestion, so that import runs with no prompt.
12. As a user with one project, I want to import a whole kubeconfig under one name, so that I keep its
    contexts together.

## Implementation Decisions

### Vocabulary

- **source kubeconfig**: a kubeconfig that `ks` does not manage, which import reads.
- **registered name**: the file name of a cluster file, without the `.yaml` suffix.
- **split mode**: import writes one cluster file for each context. This is the default.
- **single mode**: import writes one cluster file that holds every context. `--as <name>` selects it.
- **backup file**: `~/.kube/config.bkp`, the file import moves the default kubectl configuration to.

### Commands

Import adds two commands to the existing cobra command set:

- `ks import [path...]`, with flags `--as <name>` and `--yes`.
- `ks restore-kubeconfig`, with flag `--force`.

### Sources

With no path argument, import reads every path in `KUBECONFIG` when that variable is set. It uses the
platform list separator, as `kubectl` does. It reads `~/.kube/config` when `KUBECONFIG` is empty.

Path arguments replace that list.

Import refuses any path inside the `ks` configuration directory. That is how it knows the file does not
already belong to `ks`.

Import parses every source file before it writes anything. A source file that fails to parse ends the
command with an error, and import writes nothing.

Import reads a kubeconfig with the same generic YAML reader that the merge uses. Fields it does not model
survive the import.

### Contexts and names

Import collects the contexts of every source file, in the order it reads them. A context whose cluster or
user does not resolve is skipped with a warning.

For each context, import derives a suggested name. It takes the part of the context name after the last
`_`, `/`, or `:`. It then keeps only letters, digits, `-`, and `.`, and lowercases the result. It falls back
to the whole context name when that leaves nothing.

In split mode, import prompts for each context. The prompt shows the context name and the server, then the
suggested name in brackets. Enter accepts the suggestion. A typed name replaces it. A single `-` skips the
context.

The prompt reads one line from standard input. It does not use `promptui`, which the selector uses.
`promptui` appends a typed name to the suggested one when it lets the user edit the default, and returns an
empty name when it does not. A plain line prompt is also scriptable.

A name that a registered cluster already uses is refused, and the prompt asks again. Two contexts that
suggest one name behave the same way.

`--yes` skips every prompt and takes the suggested names. It appends `-2`, then `-3`, to a name that is
already taken.

Import registers nothing, and reports no error, when the user skips every context. A source kubeconfig that
holds no context is an error.

### Files import writes

In split mode, import writes one file for each context it keeps. The file holds:

- the context, with the name the source file gave it;
- the cluster that context names;
- the user that context names;
- `apiVersion: v1`, `kind: Config`, `preferences: {}`;
- `current-context`, set to that context.

Names inside the file stay as the source file wrote them. The merge renames them for display, so a
single-context file surfaces under its registered name.

A cluster or a user that several contexts share is copied into each file.

Credentials that the kubeconfig names by path stay paths. Import copies no certificate file.

In single mode, import writes one file under the name `--as` gives. It holds every context of every source
file, with `current-context` taken from the first source file that sets it. Import prompts for no name in
this mode.

Import writes files with mode `0600`, as `ks add` does.

### Takeover of the default configuration

Import takes over `~/.kube/config` only. A source file somewhere else is read, never moved.

Import performs the takeover when all of these hold:

- `~/.kube/config` exists;
- it is not a link into the `ks` configuration directory;
- `~/.kube/config.bkp` does not exist;
- the import registered a cluster that holds the active context.

The active context is the `current-context` of `~/.kube/config`. The cluster that holds it is the one
import wrote from that context, or, in single mode, the single file it wrote.

Before the move, import prints the source path, the backup path, and the cluster it will link. It states
that `ks restore-kubeconfig` undoes the move. It then waits for a yes or no answer. A no answer keeps the
registered clusters and skips the move. `--yes` skips this prompt.

The move renames `~/.kube/config` to `~/.kube/config.bkp`. Import then links `~/.kube/config` to the merged
kubeconfig of the chosen cluster, by the same step that `ks set-default` uses. It records the default
cluster, as `set-default` does.

Import skips the move, and says so, when `~/.kube/config` is already a link into the `ks` configuration
directory. That file already belongs to `ks`. Import still registers the clusters.

Import refuses to read a source path that is inside the `ks` configuration directory, and follows links to
decide. That is the check that stops a user from importing generated files.

Import skips the move, and says why, when `~/.kube/config.bkp` exists. It never overwrites an earlier
backup.

Import rebuilds every merged kubeconfig before it links, so the link never dangles.

Import ends by printing the names it registered, the backup path, and the cluster it linked.

### Restore

`ks restore-kubeconfig` moves `~/.kube/config.bkp` back to `~/.kube/config`.

It reports an error when no backup file exists.

It stops when `~/.kube/config` exists and is not a link into the `ks` configuration directory. That file is
the user's own work. `--force` overwrites it.

Restore removes a link into the `ks` configuration directory without asking. That link is one `ks` made.

Restore leaves the clusters directory alone. Nothing that import registered is lost.

It also removes the record of the default cluster, because the default configuration is no longer a `ks`
file.

## Testing Decisions

The seam is the `HOME` environment variable, as in the merged kubeconfig tests. Prompting is the one part
that a test cannot drive through that seam.

Import therefore splits in two:

- A pure part turns source files into a list of contexts with suggested names. It reads files and returns
  values. It prompts for nothing.
- A command part prompts, then writes the files.

Tests cover the pure part directly. They cover the command part in two ways: with `--yes`, which needs no
input, and with a reader that holds the answers. The command reads its prompts from a variable, so a test
replaces that variable with the answers it wants to give.

Cover this behaviour:

- A kubeconfig with three contexts produces three cluster files.
- Each file holds one context, with its cluster and its user, and sets `current-context`.
- A field the typed structs do not model survives the import.
- Suggested names come from the last segment: `gke_acme_us-east1_prod` gives `prod`.
- A context that shares a cluster with another context still gets a complete file.
- `--yes` appends a number to a name that a registered cluster already uses.
- Enter accepts the suggested name, and a typed name replaces it.
- A `-` answer skips a context, and the rest still import.
- A name that is taken makes the prompt ask again.
- A no answer to the move keeps the default config, and keeps the registered clusters.
- `--as work` writes one file that holds every context.
- Import moves `~/.kube/config` and links it to the merged kubeconfig of the active context.
- Import refuses a path inside the `ks` configuration directory.
- Import refuses a `~/.kube/config` that is already a link into that directory.
- Import skips the move when a backup file exists, and still registers the clusters.
- Import writes nothing when one source file fails to parse.
- Skipping every context registers nothing and reports no error.
- `KUBECONFIG` with two paths imports the contexts of both.
- `restore-kubeconfig` puts the file back and leaves the registered clusters.
- `restore-kubeconfig` reports an error with no backup file.
- `restore-kubeconfig` stops on a real `~/.kube/config` without `--force`.

The integration tests already load a merged kubeconfig with `kubectl`. Add one case there: after an import,
`kubectl config get-contexts` lists the names the user chose.

## Out of Scope

- A command that renames a registered cluster.
- Merging a source kubeconfig into a cluster that is already registered.
- Copying certificate files that a kubeconfig names by path.
- Undoing the registration of imported clusters. `ks delete` already removes one.
- Importing from a cluster provider directly, such as `gcloud` or `aws`.

## Further Notes

`CreateBackup` and `RecoverFromBackup` in the config package stay unused. They copy one cluster file inside
the clusters directory, which is a different job from the backup file this spec adds.

The suggested name is a guess. It is the last segment of the context name, which is the cluster name for
every provider format seen so far. The prompt exists because the guess can be wrong.
