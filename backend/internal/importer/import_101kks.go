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
	"novels-backend/internal/parserclient"
)

type Import101KksOptions struct {
	PageURL          string
	ChaptersLimit    int // 0 = all
	UploadDir        string
	Cookie           string
	UserAgent        string
	StorageStatePath string // optional: playwright storage_state JSON
	Referer          string // optional: referer for first request (see parser_101.md)
}

type Import101KksResult struct {
	NovelID       uuid.UUID
	Slug          string
	ChaptersTotal int
	ChaptersSaved int
	CoverKey      *string
}

func Import101KksResumable(
	ctx context.Context,
	db *sqlx.DB,
	opts Import101KksOptions,
	checkpoint *Checkpoint,
	onChapter func(cp *Checkpoint, chaptersSaved int) error,
) (*Import101KksResult, *Checkpoint, error) {
	if db == nil {
		return nil, nil, fmt.Errorf("db is nil")
	}
	if strings.TrimSpace(opts.PageURL) == "" {
		return nil, nil, fmt.Errorf("page url is required")
	}
	if strings.TrimSpace(opts.UploadDir) == "" {
		return nil, nil, fmt.Errorf("upload dir is required")
	}

	pc := parserclient.New()

	storagePath := ""
	if strings.TrimSpace(opts.StorageStatePath) != "" {
		storagePath = "/data/" + filepath.Base(strings.TrimSpace(opts.StorageStatePath))
	}
	resp, err := pc.Parse(ctx, parserclient.ParseRequest{
		URL:                 opts.PageURL,
		Site:                "101kks",
		ChaptersLimit:       opts.ChaptersLimit,
		UserAgent:           opts.UserAgent,
		Referer:             opts.Referer,
		CookieHeader:        opts.Cookie,
		StorageStatePath:    storagePath,
		NavigationTimeoutMS: 300_000,
		Humanize:            true,
		Locale:              "ru-RU",
		TimezoneID:          "Europe/Moscow",
		ViewportWidth:       1365,
		ViewportHeight:      768,
		HumanDelayMSMin:     220,
		HumanDelayMSMax:     950,
		CloudflareWaitMS:    12_000,
	})
	if err != nil {
		return nil, nil, err
	}

	book := resp.Book
	chRefs := resp.Book.Chapters
	total := len(chRefs)

	if checkpoint == nil {
		checkpoint = &Checkpoint{}
	}
	checkpoint.TotalChapters = total
	if checkpoint.NextIndex < 0 {
		checkpoint.NextIndex = 0
	}
	if checkpoint.NextIndex > total {
		checkpoint.NextIndex = total
	}

	if checkpoint.NovelID == uuid.Nil {
		novelID := uuid.New()
		novelSlug := generateSlug101(book.Title, novelID)

		locs := []models.CreateLocalizationRequest{
			{Lang: "ru", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
			{Lang: "en", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
			{Lang: "zh", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
			{Lang: "ja", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
			{Lang: "ko", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
			{Lang: "fr", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
			{Lang: "de", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		}

		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()

		author := "parser"
		if strings.TrimSpace(book.Author) != "" {
			author = strings.TrimSpace(book.Author)
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO novels (id, slug, translation_status, original_chapters_count, author)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO NOTHING
		`, novelID, novelSlug, models.StatusOngoing, total, author)
		if err != nil {
			return nil, nil, fmt.Errorf("insert novel: %w", err)
		}

		if err := ensureAndLinkParserGenreTag(ctx, tx, novelID); err != nil {
			return nil, nil, err
		}

		for _, l := range locs {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO novel_localizations (novel_id, lang, title, description, alt_titles)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (novel_id, lang) DO UPDATE SET
					title = EXCLUDED.title,
					description = EXCLUDED.description,
					alt_titles = EXCLUDED.alt_titles,
					updated_at = NOW()
			`, novelID, l.Lang, l.Title, l.Description, pq.Array(l.AltTitles))
			if err != nil {
				return nil, nil, fmt.Errorf("insert localization (%s): %w", l.Lang, err)
			}
		}

		if strings.HasPrefix(book.CoverURL, "http") {
			key, err := downloadAndStoreCover101(ctx, novelID, book.CoverURL, opts.UploadDir)
			if err == nil && key != "" {
				_, err = tx.ExecContext(ctx, `UPDATE novels SET cover_image_key = $1 WHERE id = $2`, key, novelID)
				if err != nil {
					return nil, nil, fmt.Errorf("update cover_image_key: %w", err)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, nil, fmt.Errorf("commit novel: %w", err)
		}

		checkpoint.NovelID = novelID
		checkpoint.Slug = novelSlug
	}

	novelID := checkpoint.NovelID
	chaptersSaved := 0
	now := time.Now().UTC()

	for i := checkpoint.NextIndex; i < total; i++ {
		if ctx.Err() != nil {
			return nil, checkpoint, ctx.Err()
		}
		if i >= len(resp.Chapters) {
			return nil, checkpoint, fmt.Errorf("parser-service returned %d chapters, expected at least %d", len(resp.Chapters), total)
		}

		ref := chRefs[i]
		ch := resp.Chapters[i]

		number := float64(i + 1)
		if ref.Number != nil && *ref.Number > 0 {
			number = float64(*ref.Number)
		}
		title := strings.TrimSpace(ch.Title)
		if title == "" {
			title = strings.TrimSpace(ref.Title)
		}
		var titlePtr *string
		if title != "" {
			titlePtr = &title
		}

		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return nil, checkpoint, fmt.Errorf("begin chapter tx: %w", err)
		}
		rolledBack := false
		rollback := func() {
			if !rolledBack {
				_ = tx.Rollback()
				rolledBack = true
			}
		}

		chapterID := uuid.New()
		err = tx.QueryRowxContext(ctx, `
			INSERT INTO chapters (id, novel_id, number, title, published_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (novel_id, number) DO UPDATE SET
				title = EXCLUDED.title,
				published_at = EXCLUDED.published_at,
				updated_at = NOW()
			RETURNING id
		`, chapterID, novelID, number, titlePtr, now).Scan(&chapterID)
		if err != nil {
			rollback()
			return nil, checkpoint, fmt.Errorf("upsert chapter #%d: %w", i+1, err)
		}

		content := strings.TrimSpace(ch.Content)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO chapter_contents (chapter_id, lang, content, word_count, source)
			VALUES
				($1, 'zh', $2, $3, 'parser'),
				($1, 'ru', $2, $3, 'parser')
			ON CONFLICT (chapter_id, lang) DO UPDATE SET
				content = EXCLUDED.content,
				word_count = EXCLUDED.word_count,
				source = EXCLUDED.source,
				updated_at = NOW()
		`, chapterID, content, len([]rune(content)))
		if err != nil {
			rollback()
			return nil, checkpoint, fmt.Errorf("upsert chapter content #%d: %w", i+1, err)
		}

		if err := tx.Commit(); err != nil {
			rollback()
			return nil, checkpoint, fmt.Errorf("commit chapter #%d: %w", i+1, err)
		}

		chaptersSaved++
		checkpoint.NextIndex = i + 1
		if onChapter != nil {
			if err := onChapter(checkpoint, chaptersSaved); err != nil {
				return nil, checkpoint, err
			}
		}
	}

	return &Import101KksResult{
		NovelID:       novelID,
		Slug:          checkpoint.Slug,
		ChaptersTotal: total,
		ChaptersSaved: chaptersSaved,
		CoverKey:      nil,
	}, checkpoint, nil
}

func Import101Kks(ctx context.Context, db *sqlx.DB, opts Import101KksOptions) (*Import101KksResult, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if strings.TrimSpace(opts.PageURL) == "" {
		return nil, fmt.Errorf("page url is required")
	}
	if strings.TrimSpace(opts.UploadDir) == "" {
		return nil, fmt.Errorf("upload dir is required")
	}

	pc := parserclient.New()

	// storage_state path must be inside parser-service container (/data mounted to ./cookies)
	storagePath := ""
	if strings.TrimSpace(opts.StorageStatePath) != "" {
		storagePath = "/data/" + filepath.Base(strings.TrimSpace(opts.StorageStatePath))
	}
	fmt.Fprintf(os.Stderr, "parser-service: parsing book+chapters site=101kks url=%s limit=%d\n", opts.PageURL, opts.ChaptersLimit)
	resp, err := pc.Parse(ctx, parserclient.ParseRequest{
		URL:              opts.PageURL,
		Site:             "101kks",
		ChaptersLimit:    opts.ChaptersLimit,
		UserAgent:        opts.UserAgent,
		Referer:          opts.Referer,
		CookieHeader:     opts.Cookie,
		StorageStatePath: storagePath,
		// 101kks can be slow / occasionally stalls behind anti-bot pages.
		// Give Playwright more time before failing.
		NavigationTimeoutMS: 300_000,
		// Human-ish mode: reduce trivial automation fingerprints + add small jitter.
		Humanize:         true,
		Locale:           "ru-RU",
		TimezoneID:       "Europe/Moscow",
		ViewportWidth:    1365,
		ViewportHeight:   768,
		HumanDelayMSMin:  220,
		HumanDelayMSMax:  950,
		// Give JS-only challenges a short chance to resolve.
		CloudflareWaitMS: 12_000,
	})
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "parser-service: parsed title=%q chapters=%d\n", resp.Book.Title, len(resp.Chapters))

	book := resp.Book
	chRefs := resp.Book.Chapters

	novelID := uuid.New()
	novelSlug := generateSlug101(book.Title, novelID)

	locs := []models.CreateLocalizationRequest{
		{Lang: "ru", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		{Lang: "en", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		{Lang: "zh", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		{Lang: "ja", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		{Lang: "ko", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		{Lang: "fr", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
		{Lang: "de", Title: book.Title, Description: strPtrOrNilOrParser101(book.Description), AltTitles: []string{"parser"}},
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	author := "parser"
	if strings.TrimSpace(book.Author) != "" {
		author = strings.TrimSpace(book.Author)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO novels (id, slug, translation_status, original_chapters_count, author)
		VALUES ($1, $2, $3, $4, $5)
	`, novelID, novelSlug, models.StatusOngoing, len(resp.Book.Chapters), author)
	if err != nil {
		return nil, fmt.Errorf("insert novel: %w", err)
	}

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
		key, err := downloadAndStoreCover101(ctx, novelID, book.CoverURL, opts.UploadDir)
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
		if i >= len(resp.Chapters) {
			return nil, fmt.Errorf("parser-service returned %d chapters, expected at least %d", len(resp.Chapters), len(chRefs))
		}
		ch := resp.Chapters[i]
		fmt.Fprintf(os.Stderr, "db: saving chapter %d/%d\n", i+1, len(chRefs))

		chapterID := uuid.New()
		number := float64(i + 1)
		if ref.Number != nil && *ref.Number > 0 {
			number = float64(*ref.Number)
		}
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

	return &Import101KksResult{
		NovelID:       novelID,
		Slug:          novelSlug,
		ChaptersTotal: len(resp.Book.Chapters),
		ChaptersSaved: chaptersSaved,
		CoverKey:      coverKey,
	}, nil
}

func generateSlug101(title string, novelID uuid.UUID) string {
	out := slug.Make(strings.TrimSpace(title))
	if out == "" {
		out = "novel-" + novelID.String()[:8]
	}
	if len(out) > 255 {
		out = out[:255]
	}
	return out
}

func strPtrOrNilOrParser101(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		v := "parser"
		return &v
	}
	return &s
}

func downloadAndStoreCover101(ctx context.Context, novelID uuid.UUID, coverURL string, uploadDir string) (string, error) {
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
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

