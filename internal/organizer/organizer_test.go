package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDirectoryCategory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "cover.jpg"), []byte("jpg"))
	mustWriteFile(t, filepath.Join(dir, "photo.png"), []byte("png"))
	mustWriteFile(t, filepath.Join(dir, "readme.txt"), []byte("txt"))

	if got := detectDirectoryCategory(dir); got != "images" {
		t.Fatalf("detectDirectoryCategory(%q) = %q, want images", dir, got)
	}
}

func TestDetectDirectoryCategorySkipsNestedCategoryDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "report.pdf"), []byte("pdf"))
	mustWriteFile(t, filepath.Join(dir, "images", "photo.jpg"), []byte("jpg"))

	if got := detectDirectoryCategory(dir); got != "documents" {
		t.Fatalf("detectDirectoryCategory(%q) = %q, want documents", dir, got)
	}
}

func TestRunMovesFilesAndDirectories(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	mustWriteFile(t, filepath.Join(source, "photo.jpg"), []byte("jpg"))
	mustWriteFile(t, filepath.Join(source, "notes.txt"), []byte("txt"))
	mustWriteFile(t, filepath.Join(source, "images", "existing.jpg"), []byte("jpg"))
	mustWriteFile(t, filepath.Join(source, "album", "cover.png"), []byte("png"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertExists(t, filepath.Join(source, "images", "photo.jpg"))
	assertExists(t, filepath.Join(source, "documents", "notes.txt"))
	assertExists(t, filepath.Join(source, "images", "album"))
	assertExists(t, filepath.Join(source, "images", "existing.jpg"))

	assertNotExists(t, filepath.Join(source, "photo.jpg"))
	assertNotExists(t, filepath.Join(source, "notes.txt"))
	assertNotExists(t, filepath.Join(source, "album"))
}

func TestRunDryRunDoesNotMoveAnything(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	pdfPath := filepath.Join(source, "file-without-extension")
	mustWriteFile(t, pdfPath, []byte("%PDF-1.4 dry-run"))

	if err := Run(Options{Source: source, DryRun: true}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertExists(t, pdfPath)
	assertNotExists(t, filepath.Join(source, "documents", "file-without-extension"))
}

func TestRunUsesMimeFallbackWhenMovingFiles(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	pdfPath := filepath.Join(source, "statement")
	mustWriteFile(t, pdfPath, []byte("%PDF-1.4 statement"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertExists(t, filepath.Join(source, "documents", "statement"))
	assertNotExists(t, pdfPath)
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %q to exist: %v", path, err)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %q not to exist, got err=%v", path, err)
	}
}
