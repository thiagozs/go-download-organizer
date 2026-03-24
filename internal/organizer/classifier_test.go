package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsCategoryDir(t *testing.T) {
	t.Parallel()

	if !IsCategoryDir("Images") {
		t.Fatal("expected Images to be recognized as category dir")
	}

	if IsCategoryDir("random-folder") {
		t.Fatal("did not expect random-folder to be recognized as category dir")
	}
}

func TestClassifyByExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "image", path: "photo.JPG", want: "images"},
		{name: "video", path: "movie.mkv", want: "videos"},
		{name: "audio", path: "track.mp3", want: "audio"},
		{name: "document", path: "book.pdf", want: "documents"},
		{name: "archive multi ext", path: "backup.tar.gz", want: "archives"},
		{name: "package", path: "app.deb", want: "packages"},
		{name: "code", path: "main.go", want: "code"},
		{name: "config", path: "config.yaml", want: "config"},
		{name: "script", path: "deploy.sh", want: "scripts"},
		{name: "devops", path: "infra.tf", want: "devops"},
		{name: "blockchain", path: "contract.sol", want: "blockchain"},
		{name: "data", path: "dataset.csv", want: "data"},
		{name: "design", path: "layout.fig", want: "design"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Classify(tt.path); got != tt.want {
				t.Fatalf("Classify(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestClassifyByMimeFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "file-without-extension")

	if err := os.WriteFile(file, []byte("%PDF-1.4 test document"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	if got := Classify(file); got != "documents" {
		t.Fatalf("Classify(%q) = %q, want documents", file, got)
	}
}

func TestFallbackCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "invoice", in: "invoice-2026", want: "documents"},
		{name: "boleto", in: "boleto-marco", want: "documents"},
		{name: "nota", in: "nota-fiscal", want: "documents"},
		{name: "camera image", in: "img_0001", want: "images"},
		{name: "camera dsc", in: "dsc_0001", want: "images"},
		{name: "unknown", in: "whatever", want: "others"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := fallbackCategory(tt.in); got != tt.want {
				t.Fatalf("fallbackCategory(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
