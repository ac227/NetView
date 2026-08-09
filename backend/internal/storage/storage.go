package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	DataDir string
}

func New(dataDir string) (*Storage, error) {
	if dataDir == "" {
		dataDir = "./data"
	}
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	s := &Storage{DataDir: abs}
	for _, d := range []string{"images", "videos", "thumbs"} {
		if err := os.MkdirAll(filepath.Join(abs, d), 0o755); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func randomName(ext string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) + ext
}

func extFor(mime string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/avif":
		return ".avif"
	case "image/svg+xml":
		return ".svg"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-matroska":
		return ".mkv"
	default:
		return ".bin"
	}
}

func (s *Storage) Dir(kind string) string {
	return filepath.Join(s.DataDir, kind)
}

func (s *Storage) SaveFile(kind string, mime string, r io.Reader) (string, int64, error) {
	dir := s.Dir(kind)
	name := randomName(extFor(mime))
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return "", 0, err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		f.Close()
		os.Remove(path)
		return "", 0, err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", 0, err
	}
	return s.RelativePath(path), n, nil
}

func (s *Storage) SaveFromReader(kind string, ext string, r io.Reader) (string, int64, error) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	name := randomName(ext)
	path := filepath.Join(s.Dir(kind), name)
	f, err := os.Create(path)
	if err != nil {
		return "", 0, err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		f.Close()
		os.Remove(path)
		return "", 0, err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", 0, err
	}
	return s.RelativePath(path), n, nil
}

func (s *Storage) Delete(path string) error {
	if path == "" {
		return nil
	}
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(s.DataDir, path)
	}
	if !s.withinDataDir(abs) {
		return fmt.Errorf("refusing to delete outside data dir: %s", path)
	}
	return os.Remove(abs)
}

func (s *Storage) withinDataDir(path string) bool {
	rel, err := filepath.Rel(s.DataDir, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (s *Storage) AbsPath(path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.DataDir, path)
}

func (s *Storage) RelativePath(abs string) string {
	rel, err := filepath.Rel(s.DataDir, abs)
	if err != nil {
		return abs
	}
	return rel
}
