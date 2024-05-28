package main

import (
	"fmt"
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

func main() {
	arg := getArgument()

	c, err := common.InitConfig()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	s, err := common.LoadState(c)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if arg == "" {
		err := s.PrintStatusCommand()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	} else {
		err := s.SetTarget(arg)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}
}
