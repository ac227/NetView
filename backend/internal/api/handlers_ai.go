package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) triggerAITag(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if !s.ai.Enabled() {
		respondError(c, http.StatusBadRequest, "AI not configured: set base_url, api_key and model in settings")
		return
	}

	ctx := c.Request.Context()
	item, err := s.media.Get(ctx, id)
	if err != nil {
		respondItemError(c, err)
		return
	}

	var result *aiResult
	if item.LocalPath != "" {
		result, err = s.runAITag(ctx, item.LocalPath, "")
	} else if item.SourceURL != "" {
		result, err = s.runAITag(ctx, "", item.SourceURL)
	} else {
		respondError(c, http.StatusBadRequest, "item has no local file or source url")
		return
	}
	if err != nil {
		respondError(c, http.StatusBadGateway, "AI tagging failed: "+err.Error())
		return
	}

	item.Title = result.Title
	item.Description = result.Description
	item.Tags = mergeTags(item.Tags, result.Tags)
	if _, err := s.media.Update(ctx, item); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to save tags")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"title": result.Title, "description": result.Description, "tags": item.Tags})
}

func (s *Server) runAITag(ctx context.Context, localPath, url string) (*aiResult, error) {
	if localPath != "" {
		r, e := s.ai.TagImage(ctx, localPath)
		if e != nil {
			return nil, e
		}
		return &aiResult{Title: r.Title, Description: r.Description, Tags: r.Tags}, nil
	}
	r, e := s.ai.TagURL(ctx, url)
	if e != nil {
		return nil, e
	}
	return &aiResult{Title: r.Title, Description: r.Description, Tags: r.Tags}, nil
}

type aiResult struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func mergeTags(existing, added []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range existing {
		t = trimLower(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	for _, t := range added {
		t = trimLower(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

func trimLower(s string) string {
	out := ""
	for _, r := range s {
		switch r {
		case ' ', '#', '-', '_', '，', '。', '、':
			out += " "
		default:
			out += string(r)
		}
	}
	return out
}
