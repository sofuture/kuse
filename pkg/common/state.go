package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// State is the loaded set of available kubeconfig targets and the current selection.
type State struct {
	targets []Link
	current Link
	config  *Config
}

// LoadState discovers available targets and the currently linked kubeconfig.
func LoadState(c *Config) (*State, error) {
	s := &State{config: c}
	if err := s.loadTargets(); err != nil {
		return nil, err
	}

	if err := s.loadCurrent(); err != nil {
		s.current.Name = "~none~"
	}

	return s, nil
}

func (s *State) loadTargets() error {
	files, err := os.ReadDir(s.config.Sources)
	if err != nil {
		return fmt.Errorf("read sources directory: %w", err)
	}

	s.targets = make([]Link, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if !isYaml(name) {
			continue
		}
		fullPath := filepath.Join(s.config.Sources, name)
		s.targets = append(s.targets, fileToLink(fullPath))
	}

	sort.Slice(s.targets, func(i, j int) bool {
		return s.targets[i].Name < s.targets[j].Name
	})

	return nil
}

func (s *State) loadCurrent() error {
	if !exists(s.config.Kubeconfig) {
		return fmt.Errorf("kubeconfig does not exist: %s", s.config.Kubeconfig)
	}
	if !isSymlink(s.config.Kubeconfig) {
		return fmt.Errorf("kubeconfig is not a symlink: %s", s.config.Kubeconfig)
	}

	link, err := os.Readlink(s.config.Kubeconfig)
	if err != nil {
		return fmt.Errorf("read kubeconfig symlink: %w", err)
	}

	// Resolve relative symlink targets against the kubeconfig directory.
	if !filepath.IsAbs(link) {
		link = filepath.Join(filepath.Dir(s.config.Kubeconfig), link)
	}
	link = filepath.Clean(link)

	s.current = fileToLink(link)
	return nil
}

func (s *State) switchLink(target string, force bool) error {
	if exists(s.config.Kubeconfig) {
		if !isSymlink(s.config.Kubeconfig) {
			if !force {
				fmt.Fprint(os.Stderr, "kubeconfig is not a symlink; overwrite anyway? [y/N]: ")
				c, err := bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil || strings.TrimSpace(strings.ToUpper(c)) != "Y" {
					return fmt.Errorf("leaving kubeconfig alone")
				}
			}
		}
		if err := os.Remove(s.config.Kubeconfig); err != nil {
			return fmt.Errorf("remove existing kubeconfig: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(s.config.Kubeconfig), 0o755); err != nil {
		return fmt.Errorf("create kubeconfig directory: %w", err)
	}

	if err := os.Symlink(target, s.config.Kubeconfig); err != nil {
		return fmt.Errorf("create symlink: %w", err)
	}

	fmt.Println("set kubeconfig to:", target)
	return nil
}

// PrintShortStatusCommand prints only the current target name (no trailing newline).
func (s *State) PrintShortStatusCommand() {
	fmt.Print(s.current.Name)
}

// PrintStatusCommand prints the current target and available targets.
func (s *State) PrintStatusCommand() {
	fmt.Println("kuse current target:", s.current.Name)
	names := make([]string, len(s.targets))
	for i, t := range s.targets {
		names[i] = t.Name
	}
	fmt.Println("available targets:", names)
}

// SetTarget switches the kubeconfig symlink to the named target.
// When force is true, a non-symlink kubeconfig is overwritten without prompting.
func (s *State) SetTarget(target string, force bool) error {
	for _, t := range s.targets {
		if t.Name == target {
			return s.switchLink(t.File, force)
		}
	}
	return fmt.Errorf("invalid target: %s", target)
}

// CurrentName returns the current target name.
func (s *State) CurrentName() string {
	return s.current.Name
}

// TargetNames returns sorted available target names.
func (s *State) TargetNames() []string {
	names := make([]string, len(s.targets))
	for i, t := range s.targets {
		names[i] = t.Name
	}
	return names
}
