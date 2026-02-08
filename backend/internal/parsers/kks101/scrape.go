package kks101

import (
	"context"
	"fmt"
	"strings"
)

type Scraper struct {
	fetcher *Fetcher
}

func NewScraper(fetcher *Fetcher) *Scraper {
	return &Scraper{fetcher: fetcher}
}

func (s *Scraper) ScrapeBook(ctx context.Context, pageURL string) (*Book, error) {
	b, err := s.fetcher.Get(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	book, err := ParseBookPage(pageURL, b)
	if err != nil {
		return nil, err
	}

	// Prefer full catalog chapters if catalog url exists.
	if strings.TrimSpace(book.CatalogURL) != "" {
		cb, err := s.fetcher.Get(ctx, book.CatalogURL)
		if err == nil {
			if full, err := ParseCatalogPage(book.CatalogURL, cb); err == nil && len(full.Chapters) > 0 {
				full.Title = firstNonEmpty(full.Title, book.Title)
				full.CoverURL = firstNonEmpty(full.CoverURL, book.CoverURL)
				full.Description = firstNonEmpty(full.Description, book.Description)
				full.Author = firstNonEmpty(full.Author, book.Author)
				full.Category = firstNonEmpty(full.Category, book.Category)
				if len(full.Tags) == 0 {
					full.Tags = book.Tags
				}
				if strings.TrimSpace(full.CatalogURL) == "" {
					full.CatalogURL = book.CatalogURL
				}
				return full, nil
			}
		}
	}

	return book, nil
}

func (s *Scraper) ScrapeChapter(ctx context.Context, chURL string) (*Chapter, error) {
	b, err := s.fetcher.Get(ctx, chURL)
	if err != nil {
		return nil, err
	}
	ch, err := ParseChapterPage(chURL, b)
	if err != nil {
		return nil, fmt.Errorf("parse chapter %s: %w", chURL, err)
	}
	return ch, nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

