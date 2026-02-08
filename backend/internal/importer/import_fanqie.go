package importer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/parsers/fanqie"
)

type ImportFanqieOptions struct {
	PageURL       string
	ChaptersLimit int // 0 = all
	UploadDir     string
	Cookie        string
	UserAgent     string
}

type ImportFanqieResult struct {
	NovelID       uuid.UUID
	Slug          string
	ChaptersTotal int
	ChaptersSaved int
	CoverKey      *string
}

func ImportFanqie(ctx context.Context, db *sqlx.DB, opts ImportFanqieOptions) (*ImportFanqieResult, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if opts.PageURL == "" {
		return nil, fmt.Errorf("page url is required")
	}
	if opts.UploadDir == "" {
		return nil, fmt.Errorf("upload dir is required")
	}

	fetcher := fanqie.NewFetcherWithOptions(opts.Cookie, opts.UserAgent)
	s := fanqie.NewScraper(fetcher)

	book, err := s.ScrapeBook(ctx, opts.PageURL)
	if err != nil {
		return nil, err
	}

	chRefs := book.Chapters
	if opts.ChaptersLimit > 0 && opts.ChaptersLimit < len(chRefs) {
		chRefs = chRefs[:opts.ChaptersLimit]
	}

	novelID := uuid.New()
	novelSlug := generateSlug(book.Title, novelID)

	// Use zh data for all localizations (at least ru must exist for the app).
	locs := []models.CreateLocalizationRequest{
		{Lang: "ru", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
		{Lang: "en", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
		{Lang: "zh", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
		{Lang: "ja", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
		{Lang: "ko", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
		{Lang: "fr", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
		{Lang: "de", Title: book.Title, Description: strPtrOrNilOrParser(book.Description), AltTitles: []string{"parser"}},
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Fill proposal-like fields that parser does not extract with sentinel "parser"
	author := "parser"
	_, err = tx.ExecContext(ctx, `
		INSERT INTO novels (id, slug, translation_status, original_chapters_count, author)
		VALUES ($1, $2, $3, $4, $5)
	`, novelID, novelSlug, models.StatusOngoing, len(book.Chapters), author)
	if err != nil {
		return nil, fmt.Errorf("insert novel: %w", err)
	}

	// Ensure placeholder genre/tag "parser" exists and link it to novel (for tests & to avoid empty metadata).
	if err := ensureAndLinkParserGenreTag(ctx, tx, novelID); err != nil {
		return nil, err
	}

	for _, l := range locs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO novel_localizations (novel_id, lang, title, description, alt_titles)
			VALUES ($1, $2, $3, $4, $5)
		`, novelID, l.Lang, l.Title, l.Description, pq.Array(l.AltTitles))
		if err != nil {
			return nil, fmt.Errorf("insert localization (%s): %w", l.Lang, err)
		}
	}

	var coverKey *string
	if strings.HasPrefix(book.CoverURL, "http") {
		key, err := downloadAndStoreCover(ctx, novelID, book.CoverURL, opts.UploadDir, fetcher)
		if err == nil && key != "" {
			_, err = tx.ExecContext(ctx, `UPDATE novels SET cover_image_key = $1 WHERE id = $2`, key, novelID)
			if err != nil {
				return nil, fmt.Errorf("update cover_image_key: %w", err)
			}
			coverKey = &key
		}
	}

	now := time.Now().UTC()
	chaptersSaved := 0
	for i, ref := range chRefs {
		ch, err := s.ScrapeChapter(ctx, ref.URL)
		if err != nil {
			return nil, err
		}

		chapterID := uuid.New()
		number := float64(i + 1)
		title := strings.TrimSpace(ch.Title)
		if title == "" {
			title = strings.TrimSpace(ref.Title)
		}
		var titlePtr *string
		if title != "" {
			titlePtr = &title
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO chapters (id, novel_id, number, title, published_at)
			VALUES ($1, $2, $3, $4, $5)
		`, chapterID, novelID, number, titlePtr, now)
		if err != nil {
			return nil, fmt.Errorf("insert chapter #%d: %w", i+1, err)
		}

		content := strings.TrimSpace(ch.Content)
		// Important: UI requests chapters in ru locale (lang=ru). For imported originals we store content both as "zh"
		// and as "ru" fallback so chapter pages work immediately after import.
		_, err = tx.ExecContext(ctx, `
			INSERT INTO chapter_contents (chapter_id, lang, content, word_count, source)
			VALUES
				($1, 'zh', $2, $3, 'parser'),
				($1, 'ru', $2, $3, 'parser')
		`, chapterID, content, len([]rune(content)))
		if err != nil {
			return nil, fmt.Errorf("insert chapter content #%d: %w", i+1, err)
		}

		chaptersSaved++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ImportFanqieResult{
		NovelID:       novelID,
		Slug:          novelSlug,
		ChaptersTotal: len(book.Chapters),
		ChaptersSaved: chaptersSaved,
		CoverKey:      coverKey,
	}, nil
}

func generateSlug(title string, novelID uuid.UUID) string {
	out := slug.Make(strings.TrimSpace(title))
	if out == "" {
		out = "novel-" + novelID.String()[:8]
	}
	if len(out) > 255 {
		out = out[:255]
	}
	return out
}

func strPtrOrNil(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func strPtrOrNilOrParser(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		v := "parser"
		return &v
	}
	return &s
}

func ensureAndLinkParserGenreTag(ctx context.Context, tx *sqlx.Tx, novelID uuid.UUID) error {
	// Genre
	var genreID uuid.UUID
	err := tx.GetContext(ctx, &genreID, `
		WITH ins AS (
			INSERT INTO genres (slug) VALUES ('parser')
			ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
			RETURNING id
		)
		SELECT id FROM ins
		UNION ALL
		SELECT id FROM genres WHERE slug = 'parser'
		LIMIT 1
	`)
	if err != nil {
		return fmt.Errorf("ensure parser genre: %w", err)
	}
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO genre_localizations (genre_id, lang, name)
		VALUES ($1, 'ru', 'parser')
		ON CONFLICT (genre_id, lang) DO NOTHING
	`, genreID)
	_, _ = tx.ExecContext(ctx, `INSERT INTO novel_genres (novel_id, genre_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, novelID, genreID)

	// Tag
	var tagID uuid.UUID
	err = tx.GetContext(ctx, &tagID, `
		WITH ins AS (
			INSERT INTO tags (slug) VALUES ('parser')
			ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
			RETURNING id
		)
		SELECT id FROM ins
		UNION ALL
		SELECT id FROM tags WHERE slug = 'parser'
		LIMIT 1
	`)
	if err != nil {
		return fmt.Errorf("ensure parser tag: %w", err)
	}
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO tag_localizations (tag_id, lang, name)
		VALUES ($1, 'ru', 'parser')
		ON CONFLICT (tag_id, lang) DO NOTHING
	`, tagID)
	_, _ = tx.ExecContext(ctx, `INSERT INTO novel_tags (novel_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, novelID, tagID)

	return nil
}

func downloadAndStoreCover(ctx context.Context, novelID uuid.UUID, coverURL string, uploadDir string, fetcher *fanqie.Fetcher) (string, error) {
	u, err := url.Parse(coverURL)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(path.Ext(u.Path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		ext = ".jpg"
	}

	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coverURL, nil)
	if err != nil {
		return "", err
	}
	// Reuse similar headers; cookie/ua are not strictly needed for cover but harmless.
	_ = fetcher
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("cover download status=%d", resp.StatusCode)
	}

	coversDir := filepath.Join(uploadDir, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		return "", err
	}

	filename := novelID.String() + ext
	filePath := filepath.Join(coversDir, filename)
	tmpPath := filePath + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, io.LimitReader(resp.Body, 20<<20)); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		return "", err
	}

	return "covers/" + filename, nil
}

