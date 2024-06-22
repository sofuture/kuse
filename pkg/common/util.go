package common

import (
	"errors"
	"os"
	"path"
	"strings"
)

type Link struct {
	Name      string
	File      string
	Extension string
}

func (l Link) String() string {
	return l.Name
}

func isYaml(filename string) bool {
	return strings.HasSuffix(filename, ".yml") || strings.HasSuffix(filename, ".yaml")
}

func trimYamlSuffix(filename string) string {
	return strings.TrimSuffix(strings.TrimSuffix(filename, ".yaml"), ".yml")
}

func isSymlink(filename string) bool {
	fi, err := os.Lstat(filename)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink == os.ModeSymlink
}

func exists(filename string) bool {
	if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}

func fileToLink(filename string) Link {
	return Link{
		Name:      trimYamlSuffix(path.Base(filename)),
		File:      filename,
		Extension: path.Ext(filename),
	}
}
