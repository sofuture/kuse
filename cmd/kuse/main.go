package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/sofuture/kuse/pkg/common"
	"os"
)

func getArgument() string {
	args := os.Args[1:]
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

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
		fmt.Println("error:", err)
		os.Exit(1)
	}

	s, err := common.LoadState(c)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if args.Short {
		s.PrintShortStatusCommand()
		os.Exit(0)
	}

	if args.Name == "" {
		err := s.PrintStatusCommand()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	} else {
		err := s.SetTarget(args.Name)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}
}
