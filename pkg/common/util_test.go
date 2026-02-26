package common

import "testing"

func TestIsYaml(t *testing.T) {
	if !isYaml("dev.yaml") {
		t.Fatalf("expected .yaml file to be recognized")
	}
	if !isYaml("dev.yml") {
		t.Fatalf("expected .yml file to be recognized")
	}
	if isYaml("dev.txt") {
		t.Fatalf("expected non-yaml file to be rejected")
	}
}

func TestFileToLinkWithUnixPath(t *testing.T) {
	link := fileToLink("/tmp/dev.yaml")
	if link.Name != "dev" {
		t.Fatalf("expected name dev, got %q", link.Name)
	}
	if link.Extension != ".yaml" {
		t.Fatalf("expected extension .yaml, got %q", link.Extension)
	}
	if link.File != "/tmp/dev.yaml" {
		t.Fatalf("expected file path preserved, got %q", link.File)
	}
}

func TestFileToLinkWithWindowsStylePath(t *testing.T) {
	link := fileToLink(`C:\Users\alice\kubeconfigs\production.yaml`)
	if link.Name != "production" {
		t.Fatalf("expected name production, got %q", link.Name)
	}
	if link.Extension != ".yaml" {
		t.Fatalf("expected extension .yaml, got %q", link.Extension)
	}
	if link.File != `C:\Users\alice\kubeconfigs\production.yaml` {
		t.Fatalf("expected file path preserved, got %q", link.File)
	}
}
