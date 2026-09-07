# kuse

`kuse` is a simple tool to manage your kubeconfig file with symlinks.

It symlinks kubeconfigs from a given directory to your main kubeconfig path.

### Releases

Versioned tags (`v*`) automatically trigger the GitHub Actions release workflow, which builds binaries for:

- Linux (`amd64`, `arm64`)
- macOS (`amd64`, `arm64`)
- Windows (`amd64`)

Each release includes platform archives and a SHA256 checksum file.

### Install

Download a release binary, or build from source:

```shell
go install github.com/sofuture/kuse/cmd/kuse@latest
# or
make
```

### How do I use it

```
Usage: kuse [--kubeconfig KUBECONFIG] [--sources SOURCES] [--short] [--version] [NAME]

Positional arguments:
  NAME                   kubeconfig target name to activate

Options:
  --kubeconfig KUBECONFIG
                         path to the active kubeconfig symlink
  --sources SOURCES      directory containing kubeconfig files
  --short                print only the current target name
  --version              print version and exit
  --help, -h             display this help and exit
```

- run `kuse` once to create a configuration file with defaults
- config lives at `$XDG_CONFIG_HOME/kuse/kuseconfig.yaml` (typically `~/.config/kuse/kuseconfig.yaml`)
  - `kubeconfig` defaults to `~/.kube/config`
  - `sources` defaults to `~/kubeconfigs` (created automatically if missing)
- use `--kubeconfig` or `--sources` to override and persist those values
- run `kuse` to show the current kubeconfig in use
- run `kuse <name>` to switch to a different target
- run `kuse --short` for prompt-friendly output (current name only, no newline)

### Example

```shell
-> % ls ~/kubeconfigs
development.yaml  production.yaml

-> % ls -l ~/.kube/config
lrwxrwxrwx /home/user/.kube/config -> /home/user/kubeconfigs/development.yaml

-> % kuse
kuse current target: development
available targets: [development production]

-> % kuse production
set kubeconfig to: /home/user/kubeconfigs/production.yaml

-> % kuse
kuse current target: production
available targets: [development production]

-> % kuse --short
production%
```

### But can't I use kubectl's built-in context management?

Sure, go for it.
