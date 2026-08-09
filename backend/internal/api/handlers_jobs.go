package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type job struct {
	ID        int64   `json:"id"`
	ItemID    int64   `json:"item_id"`
	URL       string  `json:"url"`
	Adapter   string  `json:"adapter"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress"`
	Info      string  `json:"info"`
	Error     string  `json:"error"`
}

func (s *Server) listJobs(c *gin.Context) {
	rows, err := s.db.Pool.Query(context.Background(), `
		SELECT id, item_id, url, adapter, status, progress, info, error
		FROM download_jobs ORDER BY id DESC LIMIT 100`)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list jobs")
		return
	}
	defer rows.Close()
	var jobs []job
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.ID, &j.ItemID, &j.URL, &j.Adapter, &j.Status, &j.Progress, &j.Info, &j.Error); err != nil {
			respondError(c, http.StatusInternalServerError, "scan failed")
			return
		}
		jobs = append(jobs, j)
	}
	if jobs == nil {
		jobs = []job{}
	}
	respondJSON(c, http.StatusOK, gin.H{"jobs": jobs})
}

func (s *Server) cancelJob(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	_, err := s.db.Pool.Exec(context.Background(), `
		UPDATE download_jobs SET status='cancelled', updated_at=now() WHERE id=$1 AND status IN ('pending','running')`, id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to cancel job")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"cancelled": true})
}
