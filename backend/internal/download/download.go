package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Adapter string

const (
	AdapterDirect Adapter = "direct"
	AdapterYTDLP  Adapter = "yt-dlp"
)

type Result struct {
	Path    string
	Size    int64
	Mime    string
	Adapter Adapter
	Ext     string
}

type Manager struct {
	mu       sync.Mutex
	sem      chan struct{}
	ytdlp    string
	timeout  time.Duration
}

func NewManager(maxConcurrent int, ytdlpPath string, timeout time.Duration) *Manager {
	return &Manager{
		sem:     make(chan struct{}, maxConcurrent),
		ytdlp:   ytdlpPath,
		timeout: timeout,
	}
}

// IsSupportedURL reports whether a URL looks like a direct media file.
func (m *Manager) IsSupportedURL(url string) bool {
	lower := strings.ToLower(url)
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif", ".mp4", ".webm", ".mov", ".mkv"} {
		if strings.Contains(lower, ext) {
			return true
		}
	}
	return false
}

// Acquire blocks until a download slot is free.
func (m *Manager) Acquire(ctx context.Context) error {
	select {
	case m.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a download slot.
func (m *Manager) Release() {
	<-m.sem
}

// Direct downloads a file from a direct URL into the given directory.
func (m *Manager) Direct(ctx context.Context, url, dir string) (*Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 NetView/0.1")

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := extForContentType(contentType, url)
	name := fmt.Sprintf("download_%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		os.Remove(path)
		return nil, err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return nil, err
	}
	return &Result{Path: path, Size: n, Mime: contentType, Adapter: AdapterDirect, Ext: ext}, nil
}

// YTDLP downloads media using yt-dlp (fallback for Bilibili/Douyin/YouTube etc).
func (m *Manager) YTDLP(ctx context.Context, url, dir string) (*Result, error) {
	if _, err := exec.LookPath(m.ytdlp); err != nil {
		return nil, fmt.Errorf("yt-dlp not found at %q (install it or use a direct link)", m.ytdlp)
	}
	args := []string{
		"--no-playlist",
		"--newline",
		"-o", filepath.Join(dir, "yt_%(id)s.%(ext)s"),
		"--no-warnings",
		"--no-check-certificates",
		"--user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		url,
	}
	cmd := exec.CommandContext(ctx, m.ytdlp, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %v: %s", err, firstLines(string(out), 5))
	}

	// find the newest file in dir (yt-dlp output)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var newest os.DirEntry
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), "thumb_") || strings.HasPrefix(e.Name(), "download_") {
			continue
		}
		if newest == nil {
			newest = e
			continue
		}
		infoA, _ := e.Info()
		infoB, _ := newest.Info()
		if infoA != nil && infoB != nil && infoA.ModTime().After(infoB.ModTime()) {
			newest = e
		}
	}
	if newest == nil {
		return nil, fmt.Errorf("yt-dlp produced no output file")
	}
	path := filepath.Join(dir, newest.Name())
	info, err := newest.Info()
	if err != nil {
		return nil, err
	}
	return &Result{Path: path, Size: info.Size(), Mime: mimeByExt(newest.Name()), Adapter: AdapterYTDLP, Ext: filepath.Ext(newest.Name())}, nil
}

func extForContentType(contentType, url string) string {
	lower := strings.ToLower(contentType)
	switch {
	case strings.Contains(lower, "image/jpeg"), strings.Contains(lower, "image/jpg"):
		return ".jpg"
	case strings.Contains(lower, "image/png"):
		return ".png"
	case strings.Contains(lower, "image/gif"):
		return ".gif"
	case strings.Contains(lower, "image/webp"):
		return ".webp"
	case strings.Contains(lower, "image/avif"):
		return ".avif"
	case strings.Contains(lower, "video/mp4"):
		return ".mp4"
	case strings.Contains(lower, "video/webm"):
		return ".webm"
	case strings.Contains(lower, "video/quicktime"):
		return ".mov"
	case strings.Contains(lower, "video/x-matroska"):
		return ".mkv"
	}
	lowerURL := strings.ToLower(url)
	for _, e := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif", ".mp4", ".webm", ".mov", ".mkv"} {
		if idx := strings.LastIndex(lowerURL, e); idx >= 0 {
			return e
		}
	}
	return ".bin"
}

func mimeByExt(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".avif":
		return "image/avif"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".mkv":
		return "video/x-matroska"
	}
	return "application/octet-stream"
}

func firstLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
