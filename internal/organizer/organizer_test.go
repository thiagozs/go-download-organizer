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

	got, err := detectDirectoryCategory(dir)
	if err != nil {
		t.Fatalf("detectDirectoryCategory(%q) error = %v", dir, err)
	}
	if got != "images" {
		t.Fatalf("detectDirectoryCategory(%q) = %q, want images", dir, got)
	}
}

func TestDetectDirectoryCategorySkipsNestedCategoryDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "report.pdf"), []byte("pdf"))
	mustWriteFile(t, filepath.Join(dir, "images", "photo.jpg"), []byte("jpg"))

	got, err := detectDirectoryCategory(dir)
	if err != nil {
		t.Fatalf("detectDirectoryCategory(%q) error = %v", dir, err)
	}
	if got != "documents" {
		t.Fatalf("detectDirectoryCategory(%q) = %q, want documents", dir, got)
	}
}

func TestDetectDirectoryCategoryUsesOthersOnTie(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "cover.jpg"), []byte("jpg"))
	mustWriteFile(t, filepath.Join(dir, "report.pdf"), []byte("pdf"))

	got, err := detectDirectoryCategory(dir)
	if err != nil {
		t.Fatalf("detectDirectoryCategory(%q) error = %v", dir, err)
	}
	if got != "others" {
		t.Fatalf("detectDirectoryCategory(%q) = %q, want others", dir, got)
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

func TestRunPreservesExistingFileOnNameCollision(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	mustWriteFile(t, filepath.Join(source, "images", "photo.jpg"), []byte("existing"))
	mustWriteFile(t, filepath.Join(source, "photo.jpg"), []byte("new"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertFileContent(t, filepath.Join(source, "images", "photo.jpg"), "existing")
	assertFileContent(t, filepath.Join(source, "images", "photo (1).jpg"), "new")
}

func TestRunPreservesMultiExtensionOnNameCollision(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	mustWriteFile(t, filepath.Join(source, "archives", "backup.tar.gz"), []byte("existing"))
	mustWriteFile(t, filepath.Join(source, "backup.tar.gz"), []byte("new"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertFileContent(t, filepath.Join(source, "archives", "backup.tar.gz"), "existing")
	assertFileContent(t, filepath.Join(source, "archives", "backup (1).tar.gz"), "new")
}

func TestRunPreservesHiddenFileNameOnCollision(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	mustWriteFile(t, filepath.Join(source, "config", ".env"), []byte("existing"))
	mustWriteFile(t, filepath.Join(source, ".env"), []byte("new"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertFileContent(t, filepath.Join(source, "config", ".env"), "existing")
	assertFileContent(t, filepath.Join(source, "config", ".env (1)"), "new")
}

func TestRunPreservesExistingDirectoryOnNameCollision(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	mustWriteFile(t, filepath.Join(source, "images", "album", "existing.jpg"), []byte("existing"))
	mustWriteFile(t, filepath.Join(source, "album", "new.jpg"), []byte("new"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertExists(t, filepath.Join(source, "images", "album", "existing.jpg"))
	assertExists(t, filepath.Join(source, "images", "album (1)", "new.jpg"))
}

func TestRunReusesCategoryDirectoryIgnoringCase(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	mustWriteFile(t, filepath.Join(source, "Images", "existing.jpg"), []byte("existing"))
	mustWriteFile(t, filepath.Join(source, "photo.jpg"), []byte("new"))

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertExists(t, filepath.Join(source, "Images", "photo.jpg"))
	assertNotExists(t, filepath.Join(source, "images"))
}

func TestRunSkipsSymbolicLinks(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	target := filepath.Join(source, "target-without-extension")
	mustWriteFile(t, target, []byte("unknown"))
	link := filepath.Join(source, "download-link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symbolic links are not supported: %v", err)
	}

	if err := Run(Options{Source: source}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symbolic link %q to remain, info=%v err=%v", link, info, err)
	}
}

func TestRunDoesNotUseSymbolicLinkAsCategoryDirectory(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(source, "images")); err != nil {
		t.Skipf("symbolic links are not supported: %v", err)
	}
	mustWriteFile(t, filepath.Join(source, "photo.jpg"), []byte("new"))

	if err := Run(Options{Source: source}); err == nil {
		t.Fatal("Run() error = nil, want unsafe category symlink error")
	}

	assertNotExists(t, filepath.Join(outside, "photo.jpg"))
	assertExists(t, filepath.Join(source, "photo.jpg"))
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

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	if string(content) != want {
		t.Fatalf("content of %q = %q, want %q", path, content, want)
	}
}
