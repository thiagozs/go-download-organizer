package organizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Source string
	DryRun bool
}

func Run(opts Options) error {
	entries, err := os.ReadDir(opts.Source)
	if err != nil {
		return err
	}

	total := len(entries)
	progress := newProgressRenderer(total)

	for _, entry := range entries {
		fullPath := filepath.Join(opts.Source, entry.Name())

		// IGNORA pastas já organizadas
		if entry.IsDir() && IsCategoryDir(entry.Name()) {
			progress.renderSkippedDir(entry.Name())
			continue
		}

		if entry.IsDir() {
			category := detectDirectoryCategory(fullPath)
			if err := processDirectory(fullPath, category, opts); err != nil {
				progress.finishWithError(err)
				return err
			}
			progress.advance(filepath.Base(fullPath), category, true)
			continue
		}

		if err := processFile(fullPath, opts); err != nil {
			progress.finishWithError(err)
			return err
		}
		progress.advance(filepath.Base(fullPath), Classify(fullPath), false)
	}

	progress.finish()
	return nil
}

func processFile(path string, opts Options) error {
	fileName := filepath.Base(path)
	category := Classify(path)

	targetDir := filepath.Join(opts.Source, category)
	targetPath := filepath.Join(targetDir, fileName)

	if opts.DryRun {
		return nil
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(path, targetPath)
}

func processDirectory(dir, category string, opts Options) error {
	targetDir := filepath.Join(opts.Source, category)
	targetPath := filepath.Join(targetDir, filepath.Base(dir))

	if opts.DryRun {
		return nil
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	return os.Rename(dir, targetPath)
}

func detectDirectoryCategory(dir string) string {
	count := make(map[string]int)
	total := 0

	if err := filepath.WalkDir(dir, func(path string,
		d fs.DirEntry, err error) error {
		if err != nil {
			return nil // ignora erro e continua
		}

		if d.IsDir() {
			if path != dir && IsCategoryDir(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		category := Classify(path)
		count[category]++
		total++

		return nil
	}); err != nil {
		return "others" // fallback se erro
	}

	if total == 0 {
		return "others"
	}

	// encontra categoria dominante
	max := 0
	dominant := "others"

	for cat, c := range count {
		if c > max {
			max = c
			dominant = cat
		}
	}

	return dominant
}

type progressRenderer struct {
	total     int
	processed int
	skipped   int
}

func newProgressRenderer(total int) *progressRenderer {
	return &progressRenderer{total: total}
}

func (p *progressRenderer) advance(name, category string, isDir bool) {
	p.processed++

	kindIcon := "📄"
	if isDir {
		kindIcon = "📁"
	}

	status := fmt.Sprintf("%s %s  %s -> %s", kindIcon, trimName(name), iconForCategory(category), category)
	p.render(status)
}

func (p *progressRenderer) renderSkippedDir(name string) {
	p.skipped++
	status := fmt.Sprintf("⏭️  %s  ignorado", trimName(name))
	p.render(status)
}

func (p *progressRenderer) finish() {
	status := fmt.Sprintf("✅ concluido  processados: %d  ignorados: %d", p.processed, p.skipped)
	p.render(status)
	fmt.Println()
}

func (p *progressRenderer) finishWithError(err error) {
	status := fmt.Sprintf("❌ erro  %s", trimName(err.Error()))
	p.render(status)
	fmt.Println()
}

func (p *progressRenderer) render(status string) {
	current := p.processed + p.skipped
	if p.total == 0 {
		fmt.Printf("\r[%s] 100%% (0/0) %s\033[K", strings.Repeat("=", 18), status)
		return
	}

	percent := current * 100 / p.total
	barWidth := 18
	filled := current * barWidth / p.total
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("=", filled)
	if filled < barWidth {
		bar += ">"
		bar += strings.Repeat(".", barWidth-filled-1)
	}

	if filled == barWidth {
		bar = strings.Repeat("=", barWidth)
	}

	fmt.Printf("\r[%s] %3d%% (%d/%d) %s\033[K", bar, percent, current, p.total, status)
}

func trimName(name string) string {
	const maxLen = 48
	runes := []rune(name)
	if len(runes) <= maxLen {
		return name
	}

	return string(runes[:maxLen-1]) + "…"
}

func iconForCategory(category string) string {
	switch category {
	case "images":
		return "🖼️"
	case "videos":
		return "🎬"
	case "audio":
		return "🎵"
	case "documents":
		return "📚"
	case "archives":
		return "🗜️"
	case "packages":
		return "📦"
	case "code":
		return "💻"
	case "config":
		return "⚙️"
	case "scripts":
		return "📜"
	case "devops":
		return "🛠️"
	case "blockchain":
		return "⛓️"
	case "data":
		return "📊"
	case "design":
		return "🎨"
	default:
		return "📎"
	}
}
