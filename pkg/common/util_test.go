package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExists_ReturnsTrueForBrokenSymlink(t *testing.T) {
	tempDir := t.TempDir()
	brokenLink := filepath.Join(tempDir, "broken-link")
	missingTarget := filepath.Join(tempDir, "missing.yaml")

	if err := os.Symlink(missingTarget, brokenLink); err != nil {
		t.Fatalf("Symlink returned error: %v", err)
	}

	if !exists(brokenLink) {
		t.Fatalf("exists(%q) = false, want true", brokenLink)
	}
}
