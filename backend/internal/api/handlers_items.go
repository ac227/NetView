package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netview/netview/internal/db"
	"github.com/netview/netview/internal/download"
	"github.com/netview/netview/internal/media"
	"github.com/netview/netview/internal/meta"
)

type createItemRequest struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	SourceURL   string   `json:"source_url"`
	Tags        []string `json:"tags"`
	Categories  []int64  `json:"categories"`
}

func (s *Server) listItems(c *gin.Context) {
	f := media.ListFilter{
		Keyword:  c.Query("keyword"),
		Type:     c.Query("type"),
		Status:   c.Query("status"),
		Tag:      c.Query("tag"),
		Sort:     c.Query("sort"),
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("page_size"),
	}
	if p := c.Query("page"); p != "" {
		f.Page, _ = strconv.Atoi(p)
	}
	if ps := c.Query("page_size"); ps != "" {
		f.PageSize, _ = strconv.Atoi(ps)
	}
	if fav := c.Query("favorite"); fav != "" {
		b, _ := strconv.ParseBool(fav)
		f.Favorite = &b
	}
	if cat := c.Query("category"); cat != "" {
		f.Category, _ = strconv.ParseInt(cat, 10, 64)
	}

	items, total, err := s.media.List(c.Request.Context(), f)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list items: "+err.Error())
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"items": items, "total": total})
}

func (s *Server) createItem(c *gin.Context) {
	var req createItemRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.SourceURL == "" {
		respondError(c, http.StatusBadRequest, "source_url required")
		return
	}
	if req.Type == "" {
		req.Type = detectTypeFromURL(req.SourceURL)
	}

	item := &media.Item{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		SourceURL:   req.SourceURL,
		Tags:        req.Tags,
		Categories:  req.Categories,
	}
	created, err := s.media.Create(c.Request.Context(), item)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create item: "+err.Error())
		return
	}
	respondJSON(c, http.StatusCreated, created)
}

func (s *Server) uploadItem(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "file field required")
		return
	}
	itemType := c.PostForm("type")
	if itemType == "" {
		itemType = detectTypeFromMIME(fileHeader.Header.Get("Content-Type"))
	}
	title := c.PostForm("title")

	f, err := fileHeader.Open()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to read file")
		return
	}
	defer f.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	kind := "images"
	if itemType == "video" {
		kind = "videos"
	}
	path, size, err := s.store.SaveFile(kind, contentType, f)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save file: "+err.Error())
		return
	}

	item := &media.Item{
		Type:     itemType,
		Title:    title,
		MimeType: contentType,
		Size:     size,
		LocalPath: path,
		Status:   "ready",
	}

	ctx := c.Request.Context()
	// path 已存为相对路径，处理文件时用绝对路径
	absPath := s.store.AbsPath(path)
	if itemType == "video" {
		thumb, w, h, dur, err := media.GenerateVideoThumbnail(ctx, absPath, s.store)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to extract video thumbnail: "+err.Error())
			s.store.Delete(path)
			return
		}
		item.ThumbnailPath = s.store.RelativePath(thumb)
		item.Width, item.Height, item.Duration = w, h, dur
	} else {
		thumb, w, h, err := media.GenerateImageThumbnail(absPath, s.store)
		if err == nil {
			item.ThumbnailPath = thumb
			item.Width, item.Height = w, h
		}
	}

	created, err := s.media.Create(ctx, item)
	if err != nil {
		s.store.Delete(path)
		s.store.Delete(item.ThumbnailPath)
		respondError(c, http.StatusInternalServerError, "failed to create item: "+err.Error())
		return
	}
	respondJSON(c, http.StatusCreated, created)
}

func (s *Server) fetchMeta(c *gin.Context) {
	var req struct {
		URL string `json:"url"`
	}
	if !bindJSON(c, &req) {
		return
	}
	if req.URL == "" {
		respondError(c, http.StatusBadRequest, "url required")
		return
	}
	m, err := meta.Fetch(c.Request.Context(), req.URL)
	if err != nil {
		respondError(c, http.StatusBadGateway, "failed to fetch page: "+err.Error())
		return
	}
	respondJSON(c, http.StatusOK, m)
}

func (s *Server) getItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := s.media.Get(c.Request.Context(), id)
	if err != nil {
		respondItemError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, item)
}

func (s *Server) updateItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	existing, err := s.media.Get(c.Request.Context(), id)
	if err != nil {
		respondItemError(c, err)
		return
	}
	var req createItemRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.SourceURL != "" {
		existing.SourceURL = req.SourceURL
	}
	if req.Type != "" {
		existing.Type = req.Type
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Categories != nil {
		existing.Categories = req.Categories
	}
	updated, err := s.media.Update(c.Request.Context(), existing)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update: "+err.Error())
		return
	}
	respondJSON(c, http.StatusOK, updated)
}

func (s *Server) deleteItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := s.media.Get(c.Request.Context(), id)
	if err != nil {
		respondItemError(c, err)
		return
	}
	if err := s.media.Delete(c.Request.Context(), id); err != nil {
		respondItemError(c, err)
		return
	}
	s.store.Delete(item.LocalPath)
	s.store.Delete(item.ThumbnailPath)
	respondJSON(c, http.StatusOK, gin.H{"deleted": true})
}

func (s *Server) toggleFavorite(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req struct {
		Favorite *bool `json:"favorite"`
	}
	_ = c.ShouldBindJSON(&req)
	fav := true
	if req.Favorite != nil {
		fav = *req.Favorite
	}
	if err := s.media.SetFavorite(c.Request.Context(), id, fav); err != nil {
		respondItemError(c, err)
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"favorite": fav})
}

func (s *Server) getItemFile(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := s.media.Get(c.Request.Context(), id)
	if err != nil {
		respondItemError(c, err)
		return
	}
	if item.LocalPath == "" {
		respondError(c, http.StatusNotFound, "no local file for this item")
		return
	}
	s.serveFile(c, s.store.AbsPath(item.LocalPath), item.MimeType)
}

func (s *Server) getItemThumbnail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := s.media.Get(c.Request.Context(), id)
	if err != nil {
		respondItemError(c, err)
		return
	}
	if item.ThumbnailPath == "" {
		respondError(c, http.StatusNotFound, "no thumbnail")
		return
	}
	s.serveFile(c, s.store.AbsPath(item.ThumbnailPath), "image/jpeg")
}

func (s *Server) serveFile(c *gin.Context, path, mime string) {
	file, err := os.Open(path)
	if err != nil {
		respondError(c, http.StatusNotFound, "file not found")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to stat file")
		return
	}

	c.Header("Content-Type", mime)
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		var start, end int64
		_, err := fmt.Sscanf(strings.TrimPrefix(rangeHeader, "bytes="), "%d-%d", &start, &end)
		if err != nil || start >= stat.Size() {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", stat.Size()))
			c.Status(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if end == 0 || end >= stat.Size() {
			end = stat.Size() - 1
		}
		if _, err := file.Seek(start, io.SeekStart); err != nil {
			respondError(c, http.StatusInternalServerError, "seek failed")
			return
		}
		c.Status(http.StatusPartialContent)
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, stat.Size()))
		c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
		io.CopyN(c.Writer, file, end-start+1)
		return
	}

	c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))
	http.ServeContent(c.Writer, c.Request, filepath.Base(path), stat.ModTime(), file)
}

// triggerDownload starts an async download for an item that has a source URL.
func (s *Server) triggerDownload(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	item, err := s.media.Get(ctx, id)
	if err != nil {
		respondItemError(c, err)
		return
	}
	if item.SourceURL == "" {
		respondError(c, http.StatusBadRequest, "item has no source_url")
		return
	}
	if item.Status == "downloading" {
		respondError(c, http.StatusConflict, "already downloading")
		return
	}

	adapter := download.AdapterDirect
	if !s.download.IsSupportedURL(item.SourceURL) {
		adapter = download.AdapterYTDLP
	}

	jobID, err := s.createJob(ctx, id, item.SourceURL, adapter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to create job: "+err.Error())
		return
	}
	if err := s.media.SetItemStatus(ctx, id, "downloading"); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to update status")
		return
	}
	respondJSON(c, http.StatusAccepted, gin.H{"job_id": jobID})

	go s.runJob(jobID, id)
}

func (s *Server) createJob(ctx context.Context, itemID int64, url string, adapter download.Adapter) (int64, error) {
	var jobID int64
	err := s.db.Pool.QueryRow(ctx, `
		INSERT INTO download_jobs (item_id, url, adapter, status) VALUES ($1,$2,$3,'pending')
		RETURNING id`, itemID, url, adapter).Scan(&jobID)
	return jobID, err
}

func (s *Server) runJob(jobID, itemID int64) {
	ctx := context.Background()
	set := func(status string, progress float64, errMsg string) {
		s.db.Pool.Exec(ctx, `
			UPDATE download_jobs SET status=$1, progress=$2, error=$3, updated_at=now() WHERE id=$4`,
			status, progress, errMsg, jobID)
	}
	set("running", 0.1, "")

	item, err := s.media.Get(ctx, itemID)
	if err != nil {
		set("failed", 0, "item not found")
		return
	}

	s.download.Acquire(ctx)
	defer s.download.Release()

	var res *download.Result
	if item.SourceURL != "" {
		var dctx context.Context
		var cancel context.CancelFunc
		dctx, cancel = context.WithTimeout(ctx, s.cfg.Download.Timeout)
		defer cancel()
		if s.download.IsSupportedURL(item.SourceURL) {
			res, err = s.download.Direct(dctx, item.SourceURL, s.store.Dir("videos"))
		} else {
			res, err = s.download.YTDLP(dctx, item.SourceURL, s.store.Dir("videos"))
		}
	}
	if err != nil {
		set("failed", 0, err.Error())
		s.media.SetItemStatus(ctx, itemID, "failed")
		return
	}
	if res == nil {
		set("failed", 0, "no download result")
		s.media.SetItemStatus(ctx, itemID, "failed")
		return
	}

	// 存储相对路径，避免数据目录迁移后路径失效
	item.LocalPath = s.store.RelativePath(res.Path)
	item.Size = res.Size
	item.MimeType = res.Mime
	item.Status = "downloaded"
	if item.Type == "" {
		item.Type = "video"
	}
	if res.Mime != "" && strings.HasPrefix(res.Mime, "image/") {
		item.Type = "image"
	}

	thumb, w, h, dur, terr := media.GenerateVideoThumbnail(ctx, res.Path, s.store)
	if terr == nil {
		item.ThumbnailPath = s.store.RelativePath(thumb)
		item.Width, item.Height, item.Duration = w, h, dur
	}
	// 只更新下载相关字段，绝不覆盖用户并发编辑的标题/描述/标签
	if err := s.media.ApplyDownloadResult(ctx, item); err != nil {
		set("failed", 0, err.Error())
		return
	}
	set("done", 1.0, "")
}

func detectTypeFromURL(url string) string {
	lower := strings.ToLower(url)
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif"} {
		if strings.Contains(lower, ext) {
			return "image"
		}
	}
	return "video"
}

func detectTypeFromMIME(mime string) string {
	if strings.HasPrefix(mime, "image/") {
		return "image"
	}
	return "video"
}

func respondItemError(c *gin.Context, err error) {
	if errors.Is(err, db.ErrNotFound) {
		respondError(c, http.StatusNotFound, "item not found")
		return
	}
	respondError(c, http.StatusInternalServerError, err.Error())
}
