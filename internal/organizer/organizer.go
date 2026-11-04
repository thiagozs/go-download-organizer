package organizer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Source string
	DryRun bool
	// Out receives the progress output. Defaults to os.Stdout.
	Out io.Writer
}

func Run(opts Options) error {
	entries, err := os.ReadDir(opts.Source)
	if err != nil {
		return err
	}

	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	progress := newReporter(out, len(entries), opts.DryRun)
	progress.header(opts.Source)

	for _, entry := range entries {
		fullPath := filepath.Join(opts.Source, entry.Name())

		// IGNORA pastas já organizadas
		if entry.IsDir() && IsCategoryDir(entry.Name()) {
			progress.skipped(entry.Name(), "já organizada")
			continue
		}

		if entry.IsDir() {
			if IsInProgress(entry.Name()) {
				progress.skipped(entry.Name(), "download em andamento")
				continue
			}
			category, hasInProgressDownload, err := detectDirectoryCategory(fullPath)
			if err != nil {
				progress.abort()
				return fmt.Errorf("classify directory %q: %w", fullPath, err)
			}
			if hasInProgressDownload {
				progress.skipped(entry.Name(), "contém download em andamento")
				continue
			}
			targetPath, err := processDirectory(fullPath, category, opts)
			if err != nil {
				progress.abort()
				return err
			}
			progress.recordMove(filepath.Base(fullPath), displayTarget(opts.Source, targetPath), category, true)
			continue
		}

		regular, err := isRegularEntry(entry)
		if err != nil {
			progress.abort()
			return fmt.Errorf("inspect %q: %w", fullPath, err)
		}
		if !regular {
			progress.skipped(entry.Name(), "não é arquivo regular")
			continue
		}

		if IsInProgress(entry.Name()) {
			progress.skipped(entry.Name(), "download em andamento")
			continue
		}

		category := Classify(fullPath)
		targetPath, err := processFile(fullPath, category, opts)
		if err != nil {
			progress.abort()
			return err
		}
		progress.recordMove(filepath.Base(fullPath), displayTarget(opts.Source, targetPath), category, false)
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

func detectDirectoryCategory(dir string) (string, bool, error) {
	count := make(map[string]int)
	total, tempCount := 0, 0
	hasInProgressDownload := false

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

		if IsInProgress(d.Name()) {
			hasInProgressDownload = true
			return nil
		}

		category := Classify(path)
		count[category]++
		total++
		if category == "temp" {
			tempCount++
		}

		return nil
	}); err != nil {
		return "", false, err
	}

	// Temporários só decidem a categoria quando são tudo o que a pasta tem.
	if total > count["temp"] {
		delete(count, "temp")
		total -= tempCount
	}

	if total == 0 {
		return "others", hasInProgressDownload, nil
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
		return "others", hasInProgressDownload, nil
	}

	return dominant, hasInProgressDownload, nil
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
