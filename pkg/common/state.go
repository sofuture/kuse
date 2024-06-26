package common

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

func LoadState(c *Config) (*State, error) {
	s := &State{config: c}
	err := s.loadTargets()
	if err != nil {
		return s, err
	}

	err = s.loadCurrent()
	if err != nil {
		fmt.Println(err)
		s.current.Name = "~none~"
		return s, nil
	}

	return s, nil
}

type State struct {
	targets []Link
	current Link
	config  *Config
}

func (s *State) loadTargets() error {
	files, err := os.ReadDir(s.config.Sources)
	if err != nil {
		return err
	}

	s.targets = make([]Link, 0)
	for _, file := range files {
		if isYaml(file.Name()) {
			filepath := path.Join(s.config.Sources, file.Name())
			s.targets = append(s.targets, fileToLink(filepath))
		}
	}

	return nil
}

func (s *State) loadCurrent() error {
	if exists(s.config.Kubeconfig) {
		if !isSymlink(s.config.Kubeconfig) {
			s.current = Link{}
			return errors.New("kubeconfig is not a symlink")
		}
	} else {
		return errors.New("kubeconfig does not exist")
	}

	link, err := os.Readlink(s.config.Kubeconfig)
	if err != nil {
		return err
	}

	s.current = fileToLink(link)

	return nil
}

func (s *State) switchLink(target string) error {
	if exists(s.config.Kubeconfig) {
		if !isSymlink(s.config.Kubeconfig) {
			fmt.Printf("overwrite anyway? [y/N]: ")
			c, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.TrimSpace(strings.ToUpper(c)) != "Y" || err != nil {
				return errors.New("leaving kubeconfig alone")
			}
		}

		err := os.Remove(s.config.Kubeconfig)
		if err != nil {
			return err
		}
	}

	err := os.Symlink(target, s.config.Kubeconfig)
	if err != nil {
		return err
	}

	fmt.Println("set kubeconfig to:", target)
	return nil
}

func (s *State) PrintShortStatusCommand() {
	fmt.Print(s.current.Name)
}

func (s *State) PrintStatusCommand() error {
	fmt.Println("kuse current target:", s.current.Name)
	fmt.Println("available targets:", s.targets)
	return nil
}

func (s *State) SetTarget(target string) error {
	valid := false
	filename := ""
	for _, t := range s.targets {
		if t.Name == target {
			valid = true
			filename = t.File
			break
		}
	}

	if !valid {
		return errors.New(fmt.Sprintf("invalid target: %s", target))
	}

	return s.switchLink(filename)
}
