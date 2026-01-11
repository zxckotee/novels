package handlers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"
	"github.com/rs/zerolog"
)

// XML sitemap structures
type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapIndexEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type SitemapIndex struct {
	XMLName  xml.Name            `xml:"sitemapindex"`
	XMLNS    string              `xml:"xmlns,attr"`
	Sitemaps []SitemapIndexEntry `xml:"sitemap"`
}

type SitemapHandler struct {
	novelRepo   *repository.NovelRepository
	chapterRepo *repository.ChapterRepository
	newsRepo    *repository.NewsRepository
	baseURL     string
	languages   []string
	logger      zerolog.Logger
}

func NewSitemapHandler(
	novelRepo *repository.NovelRepository,
	chapterRepo *repository.ChapterRepository,
	newsRepo *repository.NewsRepository,
	baseURL string,
	logger zerolog.Logger,
) *SitemapHandler {
	return &SitemapHandler{
		novelRepo:   novelRepo,
		chapterRepo: chapterRepo,
		newsRepo:    newsRepo,
		baseURL:     baseURL,
		languages:   []string{"ru", "en", "zh", "ja", "ko", "fr", "de"},
		logger:      logger,
	}
}

// SitemapIndex serves the main sitemap index
func (h *SitemapHandler) SitemapIndex(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Format("2006-01-02")

	index := SitemapIndex{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: []SitemapIndexEntry{
			{Loc: h.baseURL + "/sitemap-pages.xml", LastMod: now},
			{Loc: h.baseURL + "/sitemap-novels.xml", LastMod: now},
			{Loc: h.baseURL + "/sitemap-chapters.xml", LastMod: now},
			{Loc: h.baseURL + "/sitemap-news.xml", LastMod: now},
		},
	}

	h.writeXML(w, index)
}

// SitemapPages serves static pages sitemap
func (h *SitemapHandler) SitemapPages(w http.ResponseWriter, r *http.Request) {
	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []SitemapURL{},
	}

	// Static pages for each language
	staticPages := []string{"/", "/catalog", "/voting", "/collections", "/news"}

	for _, lang := range h.languages {
		for _, page := range staticPages {
			priority := "0.5"
			changeFreq := "weekly"
			if page == "/" {
				priority = "1.0"
				changeFreq = "daily"
			} else if page == "/catalog" {
				priority = "0.9"
				changeFreq = "daily"
			}

			sitemap.URLs = append(sitemap.URLs, SitemapURL{
				Loc:        fmt.Sprintf("%s/%s%s", h.baseURL, lang, page),
				ChangeFreq: changeFreq,
				Priority:   priority,
			})
		}
	}

	h.writeXML(w, sitemap)
}

// SitemapNovels serves novels sitemap
func (h *SitemapHandler) SitemapNovels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []SitemapURL{},
	}

	// TODO: Implement GetAllSlugs in NovelRepository
	// For now, use List method with large limit
	params := models.NovelListParams{
		Lang:  "ru",
		Page:  1,
		Limit: 1000, // Large limit to get all novels
		Sort:  "updated_at",
		Order: "desc",
	}
	novels, _, err := h.novelRepo.List(ctx, params)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get novels for sitemap")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for _, novel := range novels {
		for _, lang := range h.languages {
			sitemap.URLs = append(sitemap.URLs, SitemapURL{
				Loc:        fmt.Sprintf("%s/%s/novel/%s", h.baseURL, lang, novel.Slug),
				LastMod:    novel.UpdatedAt.Format("2006-01-02"),
				ChangeFreq: "weekly",
				Priority:   "0.8",
			})
		}
	}

	h.writeXML(w, sitemap)
}

// SitemapChapters serves chapters sitemap
func (h *SitemapHandler) SitemapChapters(w http.ResponseWriter, r *http.Request) {
	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []SitemapURL{},
	}

	// TODO: Implement GetAllForSitemap in ChapterRepository
	// For now, return empty sitemap
	// This method will be implemented later when needed
	h.logger.Warn().Msg("SitemapChapters: GetAllForSitemap not implemented, returning empty sitemap")

	h.writeXML(w, sitemap)
}

// SitemapNews serves news sitemap
func (h *SitemapHandler) SitemapNews(w http.ResponseWriter, r *http.Request) {
	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []SitemapURL{},
	}

	// TODO: Implement GetAllForSitemap in NewsRepository
	// For now, return empty sitemap
	// This method will be implemented later when needed
	h.logger.Warn().Msg("SitemapNews: GetAllForSitemap not implemented, returning empty sitemap")

	h.writeXML(w, sitemap)
}

// Robots.txt handler
func (h *SitemapHandler) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	robotsTxt := fmt.Sprintf(`User-agent: *
Allow: /

Sitemap: %s/sitemap.xml

# Block admin and private areas
Disallow: /admin/
Disallow: /api/
Disallow: /moderation/
Disallow: /profile/settings/
`, h.baseURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(robotsTxt))
}

func (h *SitemapHandler) writeXML(w http.ResponseWriter, data interface{}) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(data); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode XML")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
