package organizer

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var categorySet = map[string]struct{}{
	"images": {}, "videos": {}, "audio": {}, "documents": {},
	"archives": {}, "packages": {}, "code": {}, "config": {},
	"scripts": {}, "devops": {}, "blockchain": {}, "data": {},
	"design": {}, "others": {},
}

func IsCategoryDir(name string) bool {
	_, ok := categorySet[strings.ToLower(name)]
	return ok
}

// Multi-extensões (prioridade máxima)
var multiExtMap = map[string]string{
	".tar.gz":  "archives",
	".tar.bz2": "archives",
	".tar.xz":  "archives",
	".tar.zst": "archives",
}

// Extensão simples (O(1))
var extMap = map[string]string{
	// Images
	".jpg": "images", ".jpeg": "images", ".png": "images", ".gif": "images",
	".webp": "images", ".bmp": "images", ".svg": "images", ".heic": "images",

	// Videos
	".mp4": "videos", ".mkv": "videos", ".avi": "videos", ".mov": "videos",
	".wmv": "videos", ".flv": "videos", ".webm": "videos",

	// Audio
	".mp3": "audio", ".wav": "audio", ".flac": "audio", ".aac": "audio", ".ogg": "audio",

	// Documents
	".pdf": "documents", ".doc": "documents", ".docx": "documents",
	".xls": "documents", ".xlsx": "documents", ".ppt": "documents",
	".pptx": "documents", ".txt": "documents", ".rtf": "documents",
	".odt": "documents", ".md": "documents", ".epub": "documents",
	".fb2": "documents",

	// Archives
	".zip": "archives", ".rar": "archives", ".7z": "archives",
	".tar": "archives", ".gz": "archives", ".bz2": "archives",
	".xz": "archives", ".arj": "archives", ".lz": "archives",

	// Linux packages
	".deb": "packages", ".rpm": "packages", ".apk": "packages",

	// Code
	".go": "code", ".js": "code", ".ts": "code", ".py": "code",
	".java": "code", ".c": "code", ".cpp": "code", ".rs": "code",

	// Config
	".json": "config", ".yaml": "config", ".yml": "config",
	".toml": "config", ".env": "config", ".ini": "config",

	// Scripts
	".sh": "scripts", ".bash": "scripts", ".zsh": "scripts",

	// DevOps
	".tf": "devops", ".hcl": "devops",

	// Blockchain
	".sol": "blockchain", ".vy": "blockchain",

	// Data
	".csv": "data", ".parquet": "data", ".avro": "data",

	// Design
	".psd": "design", ".ai": "design", ".fig": "design",
}

func Classify(path string) string {
	fileName := filepath.Base(path)
	lower := strings.ToLower(fileName)

	// 1 Multi-extensão
	for ext, category := range multiExtMap {
		if strings.HasSuffix(lower, ext) {
			return category
		}
	}

	// 2 Extensão simples
	ext := strings.ToLower(filepath.Ext(lower))
	if category, ok := extMap[ext]; ok {
		return category
	}

	// 3 MIME (fallback inteligente)
	if mime := detectMime(path); mime != "" {
		if category := mimeToCategory(mime); category != "" {
			return category
		}
	}

	// 4 Heurística por nome
	return fallbackCategory(lower)
}

func detectMime(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return ""
	}

	return http.DetectContentType(buf[:n])
}

func mimeToCategory(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "images"
	case strings.HasPrefix(mime, "video/"):
		return "videos"
	case strings.HasPrefix(mime, "audio/"):
		return "audio"
	case strings.Contains(mime, "pdf"),
		strings.Contains(mime, "msword"),
		strings.Contains(mime, "officedocument"):
		return "documents"
	case strings.Contains(mime, "zip"),
		strings.Contains(mime, "gzip"),
		strings.Contains(mime, "tar"):
		return "archives"
	case strings.Contains(mime, "json"),
		strings.Contains(mime, "xml"):
		return "config"
	}

	return ""
}

func fallbackCategory(name string) string {
	switch {
	case strings.Contains(name, "invoice"),
		strings.Contains(name, "boleto"),
		strings.Contains(name, "nota"):
		return "documents"

	case strings.HasPrefix(name, "img_"),
		strings.HasPrefix(name, "dsc_"):
		return "images"
	}

	return "others"
}
