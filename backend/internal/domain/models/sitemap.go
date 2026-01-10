package models

import "time"

// SitemapNovel contains minimal novel info for sitemap generation
type SitemapNovel struct {
	Slug      string
	UpdatedAt time.Time
}

// SitemapChapter contains minimal chapter info for sitemap generation
type SitemapChapter struct {
	NovelSlug string
	Number    int
	UpdatedAt time.Time
}

// SitemapNews contains minimal news info for sitemap generation
type SitemapNews struct {
	Slug      string
	UpdatedAt time.Time
}
