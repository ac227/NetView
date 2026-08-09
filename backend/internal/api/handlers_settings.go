package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) getSettingsMap() (map[string]string, error) {
	rows, err := s.db.Pool.Query(context.Background(), `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, rows.Err()
}

func (s *Server) setSettings(values map[string]string) error {
	for k, v := range values {
		if _, err := s.db.Pool.Exec(context.Background(), `
			INSERT INTO settings (key, value, updated_at) VALUES ($1, $2, now())
			ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=now()`,
			k, v); err != nil {
			return err
		}
	}
	return nil
}

const (
	settingAITitle    = "ai.base_url"
	settingAIAPIKey   = "ai.api_key"
	settingAIModel    = "ai.model"
)

func (s *Server) getSettings(c *gin.Context) {
	m, err := s.getSettingsMap()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to read settings")
		return
	}
	apiKey := m[settingAIAPIKey]
	masked := ""
	if apiKey != "" {
		masked = maskKey(apiKey)
	}
	respondJSON(c, http.StatusOK, gin.H{
		"ai": gin.H{
			"base_url": m[settingAITitle],
			"api_key":  masked,
			"model":    m[settingAIModel],
			"configured": m[settingAIAPIKey] != "",
		},
		"has_password": m[keyPasswordHash] != "",
	})
}

type updateSettingsRequest struct {
	AI *struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
		Model   string `json:"model"`
	} `json:"ai"`
	Password string `json:"password"`
}

func (s *Server) updateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if !bindJSON(c, &req) {
		return
	}
	values := map[string]string{}
	if req.AI != nil {
		if req.AI.BaseURL != "" {
			values[settingAITitle] = req.AI.BaseURL
		}
		if req.AI.Model != "" {
			values[settingAIModel] = req.AI.Model
		}
		if req.AI.APIKey != "" && req.AI.APIKey != "••••••••" {
			values[settingAIAPIKey] = req.AI.APIKey
		}
	}
	if req.Password != "" {
		h, err := s.auth.HashPassword(req.Password)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to hash password")
			return
		}
		values[keyPasswordHash] = h
	}
	if len(values) > 0 {
		if err := s.setSettings(values); err != nil {
			respondError(c, http.StatusInternalServerError, "failed to save settings")
			return
		}
	}
	respondJSON(c, http.StatusOK, gin.H{"ok": true})
}

func (s *Server) getStats(c *gin.Context) {
	type row struct {
		Count int64  `db:"count"`
		Type  string `db:"type"`
	}
	var stats = struct {
		Items       int64 `json:"items"`
		Images      int64 `json:"images"`
		Videos      int64 `json:"videos"`
		Favorites   int64 `json:"favorites"`
		PendingJobs int64 `json:"pending_jobs"`
		DiskBytes   int64 `json:"disk_bytes"`
	}{}

	var total, images, videos, favs, pending int64
	if err := s.db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM items`).Scan(&total); err == nil {
		stats.Items = total
	}
	if err := s.db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM items WHERE type='image'`).Scan(&images); err == nil {
		stats.Images = images
	}
	if err := s.db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM items WHERE type='video'`).Scan(&videos); err == nil {
		stats.Videos = videos
	}
	if err := s.db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM items WHERE favorite`).Scan(&favs); err == nil {
		stats.Favorites = favs
	}
	if err := s.db.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM download_jobs WHERE status IN ('pending','running')`).Scan(&pending); err == nil {
		stats.PendingJobs = pending
	}
	if err := s.db.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(size),0) FROM items`).Scan(&stats.DiskBytes); err != nil {
		stats.DiskBytes = 0
	}
	respondJSON(c, http.StatusOK, stats)
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "••••••••"
	}
	return key[:4] + "••••••••" + key[len(key)-4:]
}
