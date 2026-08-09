package media

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/netview/netview/internal/storage"
)

const thumbMaxSize = 640

// GenerateImageThumbnail creates a thumbnail for an image file and returns its path plus dimensions.
func GenerateImageThumbnail(src string, store *storage.Storage) (string, int, int, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", 0, 0, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", 0, 0, fmt.Errorf("decode image: %w", err)
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	thumb := img
	if w > thumbMaxSize || h > thumbMaxSize {
		thumb = resize(img, w, h, thumbMaxSize)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 85}); err != nil {
		return "", 0, 0, err
	}
	path, _, err := store.SaveFromReader("thumbs", ".jpg", &buf)
	if err != nil {
		return "", 0, 0, err
	}
	return path, w, h, nil
}

// GenerateVideoThumbnail extracts a frame from a video using ffmpeg.
func GenerateVideoThumbnail(ctx context.Context, src string, store *storage.Storage) (string, int, int, int, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("ffmpeg not found: %w", err)
	}

	var duration, width, height int
	if ffprobe, perr := exec.LookPath("ffprobe"); perr == nil {
		duration, width, height = probeVideo(ffprobe, src)
	}

	// Try a few seek positions in case the video is very short or the keyframe
	// at the initial position is unavailable.
	for _, seek := range []string{"1", "0.5", "0.1", "0"} {
		outPath := filepath.Join(store.Dir("thumbs"), fmt.Sprintf("thumb_%d_%d.jpg", time.Now().UnixNano(), len(seek)))
		args := []string{"-y", "-ss", seek, "-i", src, "-vframes", "1",
			"-vf", fmt.Sprintf("scale=%d:-2", thumbMaxSize), "-pix_fmt", "yuvj420p", "-q:v", "4", outPath}
		cmd := exec.CommandContext(ctx, ffmpeg, args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			os.Remove(outPath)
			return "", 0, 0, 0, fmt.Errorf("ffmpeg: %v: %s", err, firstLines(string(out)))
		}
		if st, statErr := os.Stat(outPath); statErr == nil && st.Size() > 0 {
			return outPath, width, height, duration, nil
		}
		os.Remove(outPath)
	}
	return "", 0, 0, 0, fmt.Errorf("could not extract any frame from video")
}

func firstLines(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > 6 {
		lines = lines[:6]
	}
	return strings.Join(lines, "\n")
}

func probeVideo(ffprobe, src string) (duration, width, height int) {
	args := []string{"-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height,duration",
		"-of", "default=noprint_wrappers=1", src}
	out, err := exec.Command(ffprobe, args...).Output()
	if err != nil {
		return 0, 0, 0
	}
	for _, line := range strings.Split(string(out), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		switch key {
		case "width":
			fmt.Sscanf(val, "%d", &width)
		case "height":
			fmt.Sscanf(val, "%d", &height)
		case "duration":
			var d float64
			if _, err := fmt.Sscanf(val, "%f", &d); err == nil {
				duration = int(d)
			}
		}
	}
	return
}

func resize(img image.Image, w, h, max int) image.Image {
	scale := float64(max) / float64(maxInt(w, h))
	nw := int(float64(w) * scale)
	nh := int(float64(h) * scale)
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy := y * h / nh
		for x := 0; x < nw; x++ {
			sx := x * w / nw
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
