package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// NewsRepository handles news database operations
type NewsRepository struct {
	db *sqlx.DB
}

// NewNewsRepository creates a new news repository
func NewNewsRepository(db *sqlx.DB) *NewsRepository {
	return &NewsRepository{db: db}
}

// Create creates a new news post
func (r *NewsRepository) Create(ctx context.Context, news *models.NewsPost) error {
	query := `
		INSERT INTO news_posts (slug, title, summary, content, cover_url, category, author_id, is_pinned)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		news.Slug,
		news.Title,
		news.Summary,
		news.Content,
		news.CoverURL,
		news.Category,
		news.AuthorID,
		news.IsPinned,
	).Scan(&news.ID, &news.CreatedAt, &news.UpdatedAt)
}

// GetByID retrieves a news post by ID
func (r *NewsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.NewsPost, error) {
	var news models.NewsPost
	query := `SELECT * FROM news_posts WHERE id = $1`
	err := r.db.GetContext(ctx, &news, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &news, err
}

// GetBySlug retrieves a news post by slug
func (r *NewsRepository) GetBySlug(ctx context.Context, slug string) (*models.NewsPost, error) {
	var news models.NewsPost
	query := `SELECT * FROM news_posts WHERE slug = $1`
	err := r.db.GetContext(ctx, &news, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &news, err
}

// Update updates a news post
func (r *NewsRepository) Update(ctx context.Context, news *models.NewsPost) error {
	query := `
		UPDATE news_posts 
		SET title = $2, summary = $3, content = $4, cover_url = $5, category = $6, is_pinned = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		news.ID,
		news.Title,
		news.Summary,
		news.Content,
		news.CoverURL,
		news.Category,
		news.IsPinned,
	).Scan(&news.UpdatedAt)
}

// Delete deletes a news post
func (r *NewsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM news_posts WHERE id = $1`, id)
	return err
}

// Publish publishes a news post
func (r *NewsRepository) Publish(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE news_posts SET is_published = true, published_at = $2, updated_at = NOW() WHERE id = $1`,
		id, now)
	return err
}

// Unpublish unpublishes a news post
func (r *NewsRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE news_posts SET is_published = false, updated_at = NOW() WHERE id = $1`, id)
	return err
}

// List lists news posts with filters
func (r *NewsRepository) List(ctx context.Context, params models.NewsListParams) ([]models.NewsCard, int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if params.IsPublished == nil || *params.IsPublished {
		conditions = append(conditions, "is_published = true")
	}

	if params.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argNum))
		args = append(args, *params.Category)
		argNum++
	}

	if params.IsPinned != nil {
		conditions = append(conditions, fmt.Sprintf("is_pinned = $%d", argNum))
		args = append(args, *params.IsPinned)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM news_posts %s`, whereClause)
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// Pagination
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit

	// Main query - pinned first, then by date
	query := fmt.Sprintf(`
		SELECT id, slug, title, summary, cover_url, category, is_pinned, views_count, comments_count, published_at
		FROM news_posts
		%s
		ORDER BY is_pinned DESC, published_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	var news []models.NewsCard
	if err := r.db.SelectContext(ctx, &news, query, args...); err != nil {
		return nil, 0, err
	}

	return news, total, nil
}

// GetLatest gets latest published news for homepage
func (r *NewsRepository) GetLatest(ctx context.Context, limit int) ([]models.NewsCard, error) {
	var news []models.NewsCard
	query := `
		SELECT id, slug, title, summary, cover_url, category, is_pinned, views_count, comments_count, published_at
		FROM news_posts
		WHERE is_published = true
		ORDER BY is_pinned DESC, published_at DESC
		LIMIT $1`

	err := r.db.SelectContext(ctx, &news, query, limit)
	return news, err
}

// IncrementViews increments view count
func (r *NewsRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE news_posts SET views_count = views_count + 1 WHERE id = $1`, id)
	return err
}

// CreateLocalization creates a news localization
func (r *NewsRepository) CreateLocalization(ctx context.Context, loc *models.NewsLocalization) error {
	query := `
		INSERT INTO news_localizations (news_id, lang, title, summary, content)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (news_id, lang) DO UPDATE SET 
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			content = EXCLUDED.content,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		loc.NewsID,
		loc.Lang,
		loc.Title,
		loc.Summary,
		loc.Content,
	).Scan(&loc.ID, &loc.CreatedAt, &loc.UpdatedAt)
}

// GetLocalization gets a news localization
func (r *NewsRepository) GetLocalization(ctx context.Context, newsID uuid.UUID, lang string) (*models.NewsLocalization, error) {
	var loc models.NewsLocalization
	query := `SELECT * FROM news_localizations WHERE news_id = $1 AND lang = $2`
	err := r.db.GetContext(ctx, &loc, query, newsID, lang)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &loc, err
}

// GetLocalizations gets all localizations for a news post
func (r *NewsRepository) GetLocalizations(ctx context.Context, newsID uuid.UUID) ([]models.NewsLocalization, error) {
	var locs []models.NewsLocalization
	query := `SELECT * FROM news_localizations WHERE news_id = $1`
	err := r.db.SelectContext(ctx, &locs, query, newsID)
	return locs, err
}

// DeleteLocalization deletes a news localization
func (r *NewsRepository) DeleteLocalization(ctx context.Context, newsID uuid.UUID, lang string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM news_localizations WHERE news_id = $1 AND lang = $2`,
		newsID, lang)
	return err
}

// GetLocalizedNews gets a news post with specific localization
func (r *NewsRepository) GetLocalizedNews(ctx context.Context, slug string, lang string) (*models.NewsPost, error) {
	news, err := r.GetBySlug(ctx, slug)
	if err != nil || news == nil {
		return news, err
	}

	// Try to get localization
	loc, err := r.GetLocalization(ctx, news.ID, lang)
	if err != nil {
		return news, err
	}

	if loc != nil {
		news.Title = loc.Title
		news.Summary = loc.Summary
		news.Content = loc.Content
	}

	return news, nil
}

// GetPinnedNews gets pinned news
func (r *NewsRepository) GetPinnedNews(ctx context.Context) ([]models.NewsCard, error) {
	var news []models.NewsCard
	query := `
		SELECT id, slug, title, summary, cover_url, category, is_pinned, views_count, comments_count, published_at
		FROM news_posts
		WHERE is_published = true AND is_pinned = true
		ORDER BY published_at DESC`

	err := r.db.SelectContext(ctx, &news, query)
	return news, err
}

// SetPinned sets or unsets pinned status
func (r *NewsRepository) SetPinned(ctx context.Context, id uuid.UUID, pinned bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE news_posts SET is_pinned = $2, updated_at = NOW() WHERE id = $1`,
		id, pinned)
	return err
}
