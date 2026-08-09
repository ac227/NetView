package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/netview/netview/internal/auth"
)

const keyPasswordHash = "auth.password_hash"

type loginRequest struct {
	Password string `json:"password"`
}

type loginResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
	User    string    `json:"user"`
	Setup   bool      `json:"setup"`
}

func (s *Server) authStatus(c *gin.Context) {
	m, err := s.getSettingsMap()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to read settings")
		return
	}
	_, configured := m[keyPasswordHash]
	respondJSON(c, http.StatusOK, gin.H{
		"configured": configured,
		"needs_setup": !configured,
	})
}

func (s *Server) login(c *gin.Context) {
	var req loginRequest
	if !bindJSON(c, &req) {
		return
	}
	if req.Password == "" {
		respondError(c, http.StatusBadRequest, "password required")
		return
	}

	m, err := s.getSettingsMap()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to read settings")
		return
	}

	hash, configured := m[keyPasswordHash]
	if !configured {
		// first-run setup: adopt this password
		h, err := s.auth.HashPassword(req.Password)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "failed to hash password")
			return
		}
		if err := s.setSettings(map[string]string{keyPasswordHash: h}); err != nil {
			respondError(c, http.StatusInternalServerError, "failed to save settings")
			return
		}
		s.issueToken(c, true)
		return
	}

	if !s.auth.CheckPassword(hash, req.Password) {
		respondError(c, http.StatusUnauthorized, "wrong password")
		return
	}
	s.issueToken(c, false)
}

func (s *Server) issueToken(c *gin.Context, setup bool) {
	token, exp, err := s.auth.GenerateToken()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to issue token")
		return
	}
	respondJSON(c, http.StatusOK, loginResponse{Token: token, Expires: exp, User: auth.SharedUser, Setup: setup})
}
