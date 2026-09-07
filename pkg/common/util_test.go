package common

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsYaml(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"dev.yaml":     true,
		"prod.yml":     true,
		"DEV.YAML":     true,
		"notes.txt":    false,
		"config.yamlx": false,
		"noext":        false,
	}
	for name, want := range cases {
		if got := isYaml(name); got != want {
			t.Errorf("isYaml(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestTrimYamlSuffix(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"development.yaml": "development",
		"production.yml":   "production",
		"plain":            "plain",
		"file.YAML":        "file",
	}
	for in, want := range cases {
		if got := trimYamlSuffix(in); got != want {
			t.Errorf("trimYamlSuffix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExpandHome(t *testing.T) {
	t.Parallel()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	got, err := expandHome("~/kubeconfigs")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "kubeconfigs")
	if got != want {
		t.Fatalf("expandHome(~/kubeconfigs) = %q, want %q", got, want)
	}

	got, err = expandHome("~")
	if err != nil {
		t.Fatal(err)
	}
	if got != home {
		t.Fatalf("expandHome(~) = %q, want %q", got, home)
	}

	got, err = expandHome("/abs/path")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/abs/path" {
		t.Fatalf("expandHome(/abs/path) = %q, want /abs/path", got)
	}

	if _, err := expandHome("~someuser/cfgs"); err == nil {
		t.Fatal("expected error for user-specific home dir")
	} else if !strings.Contains(err.Error(), "cannot expand user-specific home dir") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolvePathMakesAbsolute(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	got, err := resolvePath("relative/cfgs")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "relative", "cfgs")
	if got != want {
		t.Fatalf("resolvePath(relative/cfgs) = %q, want %q", got, want)
	}
}

func TestExistsIncludesDanglingSymlink(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "missing.yaml")
	link := filepath.Join(dir, "config")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if !exists(link) {
		t.Fatal("expected dangling symlink to exist via Lstat")
	}
	if !isSymlink(link) {
		t.Fatal("expected path to be detected as symlink")
	}
}

func TestFileToLink(t *testing.T) {
	t.Parallel()
	l := fileToLink("/home/user/kubeconfigs/staging.yaml")
	if l.Name != "staging" || l.Extension != ".yaml" || l.File != "/home/user/kubeconfigs/staging.yaml" {
		t.Fatalf("unexpected link: %+v", l)
	}
}
