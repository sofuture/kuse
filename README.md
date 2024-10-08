# kuse

`kuse` is a simple tool to manage your kubeconfig file with symlinks.

It :just: symlinks various kubeconfigs from a given directory to your main kubeconfig.

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
 - run `kuse` to show the current kubeconfig in use
 - run `kuse <name>` to pick a different one

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
