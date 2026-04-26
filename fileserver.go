package main

import (
	"encoding/base64"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const localImagePrefix = "/local-image/"

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

func localImageURL(imagePath string) string {
	absPath := normalizeLocalImagePath(imagePath)
	ext := strings.ToLower(filepath.Ext(absPath))
	if !imageExtensions[ext] {
		return ""
	}
	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		return ""
	}
	token := base64.RawURLEncoding.EncodeToString([]byte(absPath))
	version := strconv.FormatInt(info.ModTime().UnixNano(), 36) + "-" + strconv.FormatInt(info.Size(), 36)
	return localImagePrefix + token + "?v=" + version
}

func localImageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, localImagePrefix) {
			next.ServeHTTP(w, r)
			return
		}
		serveLocalImage(w, r)
	})
}

func serveLocalImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	encoded := strings.TrimPrefix(r.URL.Path, localImagePrefix)
	if encoded == "" {
		http.NotFound(w, r)
		return
	}

	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	absPath := normalizeLocalImagePath(string(decoded))
	ext := strings.ToLower(filepath.Ext(absPath))
	if !imageExtensions[ext] {
		http.NotFound(w, r)
		return
	}
	if info, err := os.Stat(absPath); err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	if contentType := mime.TypeByExtension(ext); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	http.ServeFile(w, r, absPath)
}
