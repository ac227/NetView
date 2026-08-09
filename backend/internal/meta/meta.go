package meta

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Meta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Video       string `json:"video"`
	ContentType string `json:"content_type"`
}

var client = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

// Fetch retrieves page metadata from a URL by parsing HTML meta/OG tags.
// The response body is limited to 5MB.
func Fetch(ctx context.Context, url string) (*Meta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36 NetView/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body := io.LimitReader(resp.Body, 5<<20)
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	m := &Meta{ContentType: strings.TrimSpace(resp.Header.Get("Content-Type"))}
	if m.ContentType == "" || strings.HasPrefix(m.ContentType, "text/html") {
		parseHTML(string(data), m)
	}
	return m, nil
}

func parseHTML(data string, m *Meta) {
	doc, err := html.Parse(strings.NewReader(data))
	if err != nil {
		return
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if m.Title == "" {
					m.Title = strings.TrimSpace(nodeText(n))
				}
			case "meta":
				parseMeta(n, m)
			case "img":
				if m.Image == "" {
					for _, a := range n.Attr {
						if a.Key == "src" && strings.HasPrefix(a.Val, "http") {
							m.Image = a.Val
						}
					}
				}
			case "video":
				if m.Video == "" {
					for _, a := range n.Attr {
						if a.Key == "src" && strings.HasPrefix(a.Val, "http") {
							m.Video = a.Val
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
}

func parseMeta(n *html.Node, m *Meta) {
	props := map[string]string{}
	for _, a := range n.Attr {
		switch a.Key {
		case "property", "name", "content":
			if a.Key == "content" {
				props["content"] = a.Val
			} else {
				props[a.Key] = strings.ToLower(a.Val)
			}
		}
	}
	prop := props["property"]
	if prop == "" {
		prop = props["name"]
	}
	switch prop {
	case "og:title", "twitter:title":
		if m.Title == "" {
			m.Title = strings.TrimSpace(props["content"])
		}
	case "og:description", "twitter:description", "description":
		if m.Description == "" {
			m.Description = strings.TrimSpace(props["content"])
		}
	case "og:image", "twitter:image":
		if m.Image == "" {
			m.Image = strings.TrimSpace(props["content"])
		}
	case "og:video":
		if m.Video == "" {
			m.Video = strings.TrimSpace(props["content"])
		}
	}
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}
