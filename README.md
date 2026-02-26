# kuse

`kuse` is a simple tool to manage your kubeconfig file with symlinks.

It :just: symlinks various kubeconfigs from a given directory to your main kubeconfig.

### Releases

Versioned tags (`v*`) automatically trigger the GitHub Actions release workflow, which builds binaries for:

- Linux (`amd64`, `arm64`)
- macOS (`amd64`, `arm64`)
- Windows (`amd64`)

Each release includes platform archives and a SHA256 checksum file.

### How do I use it

```
Usage: kuse [--kubeconfig KUBECONFIG] [--sources SOURCES] [--short] [NAME]

Positional arguments:
  NAME

Options:
  --kubeconfig KUBECONFIG
  --sources SOURCES
  --short
  --help, -h
```
 
 - just run `kuse`, it will create a configuration file
 - in the configuration file `XDG_CONFIG_HOME/kuse/kuseconfig.yaml` you can set the location  of your kubeconfig (defaults to `~/.kube/config`) and your source kubeconfig directory (defaults to `~/kubeconfigs`)
   - you can also use `--kubeconfig` or `--sources` at any time to set those values
   - passing only one of those flags updates just that key and preserves the other existing value in config
 - run `kuse` to show the current kubeconfig in use
 - run `kuse <name>` to pick a different one
 - `kuse --short` prints only the current target token to stdout (for prompts/scripts); any warnings/errors go to stderr

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
set kubeconfig to: /home/jz/kubeconfigs/production.yaml

-> % kuse
kuse current target: production
available targets: [development production]

-> % kuse -short
production%   
```

### But can't I use kubectl's built in context management?

Sure, go for it.

### Safety notes

- `kuse` expects the configured kubeconfig path to be a symlink.
- If a regular file exists at that location, `kuse` will prompt before removing it and replacing it with a symlink.
- Keep backups of important kubeconfig files if you are unsure about your local setup.

### Development

```shell
make
go test ./...
go vet ./...
go install golang.org/x/vuln/cmd/govulncheck@latest
$(go env GOPATH)/bin/govulncheck ./...
```

Optional static security scan:

```shell
go install github.com/securego/gosec/v2/cmd/gosec@latest
$(go env GOPATH)/bin/gosec ./...
```
