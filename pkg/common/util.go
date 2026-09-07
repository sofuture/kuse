package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Link represents a named kubeconfig file under the sources directory.
type Link struct {
	Name      string
	File      string
	Extension string
}

func (l Link) String() string {
	return l.Name
}

func isYaml(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".yml" || ext == ".yaml"
}

func trimYamlSuffix(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".yaml" || ext == ".yml" {
		return strings.TrimSuffix(filename, filepath.Ext(filename))
	}
	return filename
}

func isSymlink(filename string) bool {
	fi, err := os.Lstat(filename)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// exists reports whether path exists, including as a dangling symlink.
func exists(filename string) bool {
	_, err := os.Lstat(filename)
	return err == nil
}

func fileToLink(filename string) Link {
	base := filepath.Base(filename)
	return Link{
		Name:      trimYamlSuffix(base),
		File:      filename,
		Extension: filepath.Ext(base),
	}
}

// expandHome expands a leading "~" or "~/" using the current user's home directory.
// Other "~..." forms (e.g. "~otheruser") are rejected, matching go-homedir's behavior.
func expandHome(p string) (string, error) {
	if p == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	if strings.HasPrefix(p, "~") {
		return "", fmt.Errorf("cannot expand user-specific home dir")
	}
	return p, nil
}

// resolvePath expands "~"/"~/" and converts the result to an absolute path.
func resolvePath(p string) (string, error) {
	expanded, err := expandHome(p)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return abs, nil
}
