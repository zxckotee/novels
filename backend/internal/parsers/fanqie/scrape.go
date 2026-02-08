package fanqie

import (
	"context"
	"fmt"
)

type Scraper struct {
	fetcher *Fetcher
}

func NewScraper(fetcher *Fetcher) *Scraper {
	if fetcher == nil {
		fetcher = NewFetcher()
	}
	return &Scraper{fetcher: fetcher}
}

func (s *Scraper) ScrapeBook(ctx context.Context, pageURL string) (*Book, error) {
	b, err := s.fetcher.Get(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	return ParseBookPage(pageURL, b)
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

