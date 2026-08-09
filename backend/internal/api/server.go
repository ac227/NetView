package api

import (
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/netview/netview/internal/ai"
	"github.com/netview/netview/internal/auth"
	"github.com/netview/netview/internal/config"
	"github.com/netview/netview/internal/db"
	"github.com/netview/netview/internal/download"
	"github.com/netview/netview/internal/media"
	"github.com/netview/netview/internal/storage"
	"github.com/netview/netview/internal/web"
)

type Server struct {
	cfg      *config.Config
	db       *db.DB
	store    *storage.Storage
	auth     *auth.Manager
	media    *media.Repo
	download *download.Manager
	ai       *ai.Client
}

func NewServer(cfg *config.Config, database *db.DB, store *storage.Storage) *Server {
	return &Server{
		cfg:      cfg,
		db:       database,
		store:    store,
		auth:     auth.NewManager(cfg.Auth.JWTSecret, cfg.Auth.TokenTTL),
		media:    media.NewRepo(database.Pool),
		download: download.NewManager(cfg.Download.MaxConcurrent, cfg.Download.YTDLP, cfg.Download.Timeout),
		ai:       ai.NewClient(ai.Config{BaseURL: cfg.AI.BaseURL, APIKey: cfg.AI.APIKey, Model: cfg.AI.Model}),
	}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(corsMiddleware())

	// public routes
	authGroup := r.Group("/api/auth")
	{
		authGroup.GET("/status", s.authStatus)
		authGroup.POST("/login", s.login)
	}

	api := r.Group("/api")
	api.Use(s.requireAuth())

	api.GET("/items", s.listItems)
	api.POST("/items", s.createItem)
	api.POST("/items/upload", s.uploadItem)
	api.POST("/items/fetch-meta", s.fetchMeta)
	api.GET("/items/:id", s.getItem)
	api.PUT("/items/:id", s.updateItem)
	api.DELETE("/items/:id", s.deleteItem)
	api.POST("/items/:id/favorite", s.toggleFavorite)
	api.POST("/items/:id/download", s.triggerDownload)
	api.POST("/items/:id/ai-tag", s.triggerAITag)
	api.GET("/items/:id/file", s.getItemFile)
	api.GET("/items/:id/thumbnail", s.getItemThumbnail)

	api.GET("/tags", s.listTags)
	api.GET("/categories", s.listCategories)
	api.POST("/categories", s.createCategory)
	api.PUT("/categories/:id", s.updateCategory)
	api.DELETE("/categories/:id", s.deleteCategory)
	api.GET("/download/jobs", s.listJobs)
	api.POST("/download/jobs/:id/cancel", s.cancelJob)
	api.GET("/settings", s.getSettings)
	api.PUT("/settings", s.updateSettings)
	api.GET("/system/stats", s.getStats)

	s.mountFrontend(r)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (s *Server) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			// 媒体文件（<img>/<video>）无法携带 Authorization 头，允许用 ?token= 参数。
			token = c.Query("token")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := s.auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user", claims.User)
		c.Next()
	}
}

func (s *Server) settings() (map[string]string, error) {
	return s.getSettingsMap()
}

func (s *Server) mountFrontend(r *gin.Engine) {
	distFS, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(distFS))
	index, _ := fs.ReadFile(distFS, "index.html")

	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}
		if _, err := fs.Stat(distFS, strings.TrimPrefix(c.Request.URL.Path, "/")); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		// SPA fallback: return index.html for client-side routes.
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
		c.Abort()
	})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func bindJSON(c *gin.Context, v interface{}) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		respondError(c, http.StatusBadRequest, "invalid json: "+err.Error())
		return false
	}
	return true
}

func respondJSON(c *gin.Context, status int, v interface{}) {
	c.JSON(status, v)
}

func respondError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
