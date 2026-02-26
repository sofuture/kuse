package main

import (
	"errors"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/sofuture/kuse/pkg/common"
	"os"
)

var args struct {
	Name       string `arg:"positional"`
	Kubeconfig string
	Sources    string
	Short      bool
}

func main() {
	arg.MustParse(&args)

	c, err := common.InitConfig(args.Kubeconfig, args.Sources)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	s, err := common.LoadState(c)
	if err != nil {
		var stateWarning *common.StateWarning
		if errors.As(err, &stateWarning) {
			fmt.Fprintln(os.Stderr, "warning:", stateWarning)
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}

	if args.Short {
		s.PrintShortStatusCommand()
		os.Exit(0)
	}

	if args.Name == "" {
		err := s.PrintStatusCommand()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	} else {
		err := s.SetTarget(args.Name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}
}
