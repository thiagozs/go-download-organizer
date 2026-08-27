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
	progress := newProgressRenderer(total, opts.DryRun)

	for _, entry := range entries {
		fullPath := filepath.Join(opts.Source, entry.Name())

		// IGNORA pastas já organizadas
		if entry.IsDir() && IsCategoryDir(entry.Name()) {
			progress.renderSkippedDir(entry.Name())
			continue
		}

		if entry.IsDir() {
			category, err := detectDirectoryCategory(fullPath)
			if err != nil {
				progress.stop()
				return fmt.Errorf("classify directory %q: %w", fullPath, err)
			}
			targetPath, err := processDirectory(fullPath, category, opts)
			if err != nil {
				progress.stop()
				return err
			}
			progress.advance(filepath.Base(fullPath), displayTarget(opts.Source, targetPath), category, true)
			continue
		}

		regular, err := isRegularEntry(entry)
		if err != nil {
			progress.stop()
			return fmt.Errorf("inspect %q: %w", fullPath, err)
		}
		if !regular {
			progress.renderSkippedItem(entry.Name())
			continue
		}

		category := Classify(fullPath)
		targetPath, err := processFile(fullPath, category, opts)
		if err != nil {
			progress.stop()
			return err
		}
		progress.advance(filepath.Base(fullPath), displayTarget(opts.Source, targetPath), category, false)
	}

	progress.finish()
	return nil
}

func processFile(path, category string, opts Options) (string, error) {
	fileName := filepath.Base(path)
	targetDir, err := resolveCategoryDir(opts.Source, category)
	if err != nil {
		return "", err
	}
	targetPath, err := availableTargetPath(filepath.Join(targetDir, fileName), false)
	if err != nil {
		return "", err
	}

	if opts.DryRun {
		return targetPath, nil
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create category directory %q: %w", targetDir, err)
	}

	if err := os.Rename(path, targetPath); err != nil {
		return "", fmt.Errorf("move %q to %q: %w", path, targetPath, err)
	}
	return targetPath, nil
}

func processDirectory(dir, category string, opts Options) (string, error) {
	targetDir, err := resolveCategoryDir(opts.Source, category)
	if err != nil {
		return "", err
	}
	targetPath, err := availableTargetPath(filepath.Join(targetDir, filepath.Base(dir)), true)
	if err != nil {
		return "", err
	}

	if opts.DryRun {
		return targetPath, nil
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create category directory %q: %w", targetDir, err)
	}

	if err := os.Rename(dir, targetPath); err != nil {
		return "", fmt.Errorf("move %q to %q: %w", dir, targetPath, err)
	}
	return targetPath, nil
}

func detectDirectoryCategory(dir string) (string, error) {
	count := make(map[string]int)
	total := 0

	if err := filepath.WalkDir(dir, func(path string,
		d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path != dir && IsCategoryDir(filepath.Base(path)) {
				return filepath.SkipDir
			}
			return nil
		}

		regular, err := isRegularEntry(d)
		if err != nil {
			return err
		}
		if !regular {
			return nil
		}

		category := Classify(path)
		count[category]++
		total++

		return nil
	}); err != nil {
		return "", err
	}

	if total == 0 {
		return "others", nil
	}

	max := 0
	dominant := "others"
	tied := false

	for cat, c := range count {
		if c > max {
			max = c
			dominant = cat
			tied = false
		} else if c == max {
			tied = true
		}
	}

	if tied {
		return "others", nil
	}

	return dominant, nil
}

func isRegularEntry(entry fs.DirEntry) (bool, error) {
	if entry.Type()&os.ModeSymlink != 0 {
		return false, nil
	}

	info, err := entry.Info()
	if err != nil {
		return false, err
	}
	return info.Mode().IsRegular(), nil
}

func resolveCategoryDir(source, category string) (string, error) {
	exact := filepath.Join(source, category)
	if info, err := os.Lstat(exact); err == nil {
		if info.IsDir() {
			return exact, nil
		}
		return "", fmt.Errorf("category destination %q exists and is not a directory", exact)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect category directory %q: %w", exact, err)
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return "", fmt.Errorf("read source directory %q: %w", source, err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.EqualFold(entry.Name(), category) {
			return filepath.Join(source, entry.Name()), nil
		}
	}

	return exact, nil
}

func availableTargetPath(targetPath string, isDir bool) (string, error) {
	if _, err := os.Lstat(targetPath); os.IsNotExist(err) {
		return targetPath, nil
	} else if err != nil {
		return "", fmt.Errorf("inspect destination %q: %w", targetPath, err)
	}

	dir := filepath.Dir(targetPath)
	base := filepath.Base(targetPath)
	stem, ext := base, ""
	if !isDir {
		stem, ext = splitExtension(base)
	}

	for suffix := 1; ; suffix++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", stem, suffix, ext))
		if _, err := os.Lstat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("inspect destination %q: %w", candidate, err)
		}
	}
}

func splitExtension(name string) (string, string) {
	lower := strings.ToLower(name)
	ext := filepath.Ext(name)
	for multiExt := range multiExtMap {
		if strings.HasSuffix(lower, multiExt) && len(multiExt) > len(ext) {
			ext = name[len(name)-len(multiExt):]
		}
	}
	stem := strings.TrimSuffix(name, ext)
	if stem == "" {
		return name, ""
	}
	return stem, ext
}

func displayTarget(source, targetPath string) string {
	relative, err := filepath.Rel(source, targetPath)
	if err != nil {
		return targetPath
	}
	return filepath.ToSlash(relative)
}

type progressRenderer struct {
	total     int
	processed int
	skipped   int
	inline    bool
}

func newProgressRenderer(total int, dryRun bool) *progressRenderer {
	info, err := os.Stdout.Stat()
	inline := err == nil && info.Mode()&os.ModeCharDevice != 0 && !dryRun
	return &progressRenderer{total: total, inline: inline}
}

func (p *progressRenderer) advance(name, target, category string, isDir bool) {
	p.processed++

	kindIcon := "📄"
	if isDir {
		kindIcon = "📁"
	}

	status := fmt.Sprintf("%s %s  %s -> %s", kindIcon, trimName(name), iconForCategory(category), target)
	p.render(status)
}

func (p *progressRenderer) renderSkippedDir(name string) {
	p.skipped++
	status := fmt.Sprintf("⏭️  %s  ignorado", trimName(name))
	p.render(status)
}

func (p *progressRenderer) renderSkippedItem(name string) {
	p.skipped++
	status := fmt.Sprintf("⏭️  %s  ignorado (não é arquivo regular)", trimName(name))
	p.render(status)
}

func (p *progressRenderer) finish() {
	status := fmt.Sprintf("✅ concluido  processados: %d  ignorados: %d", p.processed, p.skipped)
	p.render(status)
	if p.inline {
		fmt.Println()
	}
}

func (p *progressRenderer) stop() {
	if p.inline {
		fmt.Println()
	}
}

func (p *progressRenderer) render(status string) {
	current := p.processed + p.skipped
	prefix := ""
	suffix := "\n"
	if p.inline {
		prefix = "\r"
		suffix = "\033[K"
	}
	if p.total == 0 {
		fmt.Printf("%s[%s] 100%% (0/0) %s%s", prefix, strings.Repeat("=", 18), status, suffix)
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

	fmt.Printf("%s[%s] %3d%% (%d/%d) %s%s", prefix, bar, percent, current, p.total, status, suffix)
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
