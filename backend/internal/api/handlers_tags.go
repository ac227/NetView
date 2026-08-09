package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netview/netview/internal/db"
)

func (s *Server) listTags(c *gin.Context) {
	tags, err := s.media.ListTags(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list tags")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"tags": tags})
}

type category struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
	Sort     int    `json:"sort"`
	Count    int    `json:"count"`
}

func (s *Server) listCategories(c *gin.Context) {
	rows, err := s.db.Pool.Query(context.Background(), `
		SELECT c.id, c.name, c.parent_id, c.sort, COUNT(ic.item_id)
		FROM categories c
		LEFT JOIN item_categories ic ON ic.category_id = c.id
		GROUP BY c.id ORDER BY c.sort ASC, c.name ASC`)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list categories")
		return
	}
	defer rows.Close()

	var cats []category
	for rows.Next() {
		var cat category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.ParentID, &cat.Sort, &cat.Count); err != nil {
			respondError(c, http.StatusInternalServerError, "scan failed")
			return
		}
		cats = append(cats, cat)
	}
	if cats == nil {
		cats = []category{}
	}
	respondJSON(c, http.StatusOK, gin.H{"categories": cats})
}

type categoryRequest struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
	Sort     int    `json:"sort"`
}

func (s *Server) createCategory(c *gin.Context) {
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		respondError(c, http.StatusBadRequest, "name required")
		return
	}
	var id int64
	err := s.db.Pool.QueryRow(context.Background(), `
		INSERT INTO categories (name, parent_id, sort) VALUES ($1,$2,$3) RETURNING id`,
		req.Name, req.ParentID, req.Sort).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			respondError(c, http.StatusConflict, "category with this name already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to create category")
		return
	}
	respondJSON(c, http.StatusCreated, gin.H{"id": id})
}

func (s *Server) updateCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	_, err := s.db.Pool.Exec(context.Background(), `
		UPDATE categories SET name=$1, parent_id=$2, sort=$3 WHERE id=$4`,
		req.Name, req.ParentID, req.Sort, id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			respondError(c, http.StatusNotFound, "category not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "failed to update category")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"updated": true})
}

func (s *Server) deleteCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	_, err := s.db.Pool.Exec(context.Background(), `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to delete category")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"deleted": true})
}
