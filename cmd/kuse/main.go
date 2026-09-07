package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
	"github.com/sofuture/kuse/pkg/common"
)

// version is set at build time via -ldflags.
var version = "dev"

type cliArgs struct {
	Name       string `arg:"positional" help:"kubeconfig target name to activate"`
	Kubeconfig string `arg:"--kubeconfig" help:"path to the active kubeconfig symlink"`
	Sources    string `arg:"--sources" help:"directory containing kubeconfig files"`
	Short      bool   `arg:"--short" help:"print only the current target name"`
	Force      bool   `arg:"--force,-f" help:"overwrite a non-symlink kubeconfig without prompting"`
}

func (cliArgs) Description() string {
	return "kuse manages your kubeconfig via symlinks to named configs in a sources directory."
}

func (cliArgs) Version() string {
	return version
}

func main() {
	var opts cliArgs
	arg.MustParse(&opts)

	c, err := common.InitConfig(opts.Kubeconfig, opts.Sources)
	if err != nil {
		fatal(err)
	}

	s, err := common.LoadState(c)
	if err != nil {
		fatal(err)
	}

	if opts.Short {
		s.PrintShortStatusCommand()
		return
	}

	if opts.Name == "" {
		s.PrintStatusCommand()
		return
	}

	if err := s.SetTarget(opts.Name, opts.Force); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
