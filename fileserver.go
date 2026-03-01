package main

import (
	"encoding/base64"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type LocalImageHandler struct{}

var imageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".bmp":  true,
	".ico":  true,
}

func (h *LocalImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prefix := "/local-image/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		// Not our route — return nil-like behavior so Wails handles it
		http.NotFound(w, r)
		return
	}

	encoded := strings.TrimPrefix(r.URL.Path, prefix)
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		http.Error(w, "invalid path encoding", http.StatusBadRequest)
		return
	}
	absPath := string(decoded)

	// Validate it's an image file by extension
	ext := strings.ToLower(filepath.Ext(absPath))
	if !imageExtensions[ext] {
		http.Error(w, "not an image file", http.StatusForbidden)
		return
	}

	// Check file exists
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	http.ServeFile(w, r, absPath)
}
