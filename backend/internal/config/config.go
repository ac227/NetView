package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Storage      StorageConfig
	Auth         AuthConfig
	AI           AIConfig
	Download     DownloadConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	DSN string
}

type StorageConfig struct {
	DataDir string
}

type AuthConfig struct {
	JWTSecret string
	TokenTTL  time.Duration
}

type AIConfig struct {
	Enabled  bool
	BaseURL  string
	APIKey   string
	Model    string
}

type DownloadConfig struct {
	YTDLP     string
	MaxConcurrent int
	Timeout   time.Duration
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("NETVIEW_HOST", "0.0.0.0"),
			Port: getEnvInt("NETVIEW_PORT", 8080),
		},
		Database: DatabaseConfig{
			DSN: getEnv("NETVIEW_DB_DSN", "postgres://netview:netview_dev@localhost:5432/netview"),
		},
		Storage: StorageConfig{
			// 默认跟随可执行文件所在目录，避免从不同工作目录启动导致数据“消失”。
			DataDir: getEnv("NETVIEW_DATA_DIR", defaultDataDir()),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("NETVIEW_JWT_SECRET", "netview-dev-secret-change-me"),
			TokenTTL:  getEnvDuration("NETVIEW_TOKEN_TTL", 24*time.Hour),
		},
		AI: AIConfig{
			Enabled: getEnvBool("NETVIEW_AI_ENABLED", false),
			BaseURL: getEnv("NETVIEW_AI_BASE_URL", "https://api.openai.com/v1"),
			APIKey:  getEnv("NETVIEW_AI_API_KEY", ""),
			Model:   getEnv("NETVIEW_AI_MODEL", "gpt-4o-mini"),
		},
		Download: DownloadConfig{
			YTDLP:        getEnv("NETVIEW_YTDLP_PATH", "yt-dlp"),
			MaxConcurrent: getEnvInt("NETVIEW_DOWNLOAD_MAX_CONCURRENT", 2),
			Timeout:      getEnvDuration("NETVIEW_DOWNLOAD_TIMEOUT", 2*time.Hour),
		},
	}
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// defaultDataDir returns "<可执行文件目录>/data"。用 os.Executable 定位，
// 保证无论从哪个工作目录启动，数据都落在同一位置。
func defaultDataDir() string {
	if exe, err := os.Executable(); err == nil {
		if dir := filepath.Dir(exe); dir != "" {
			return filepath.Join(dir, "data")
		}
	}
	return "./data"
}
