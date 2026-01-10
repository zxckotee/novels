package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// NovelRepository репозиторий для работы с новеллами
type NovelRepository struct {
	db *sqlx.DB
}

// NewNovelRepository создает новый NovelRepository
func NewNovelRepository(db *sqlx.DB) *NovelRepository {
	return &NovelRepository{db: db}
}

// List получает список новелл с фильтрацией и пагинацией
func (r *NovelRepository) List(ctx context.Context, params models.NovelListParams) ([]models.NovelWithLocalization, int, error) {
	// Базовый запрос
	baseQuery := `
		FROM novels n
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = $1
	`
	
	args := []interface{}{params.Lang}
	argIndex := 2
	whereConditions := []string{}

	// Фильтр по статусу
	if len(params.Status) > 0 {
		placeholders := make([]string, len(params.Status))
		for i, status := range params.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf("n.translation_status IN (%s)", strings.Join(placeholders, ",")))
	}

	// Фильтр по жанрам
	if len(params.Genres) > 0 {
		placeholders := make([]string, len(params.Genres))
		for i, genre := range params.Genres {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, genre)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf(`
			EXISTS (
				SELECT 1 FROM novel_genres ng
				JOIN genres g ON ng.genre_id = g.id
				WHERE ng.novel_id = n.id AND g.slug IN (%s)
			)
		`, strings.Join(placeholders, ",")))
	}

	// Фильтр по тегам
	if len(params.Tags) > 0 {
		placeholders := make([]string, len(params.Tags))
		for i, tag := range params.Tags {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, tag)
			argIndex++
		}
		whereConditions = append(whereConditions, fmt.Sprintf(`
			EXISTS (
				SELECT 1 FROM novel_tags nt
				JOIN tags t ON nt.tag_id = t.id
				WHERE nt.novel_id = n.id AND t.slug IN (%s)
			)
		`, strings.Join(placeholders, ",")))
	}

	// Фильтр по году
	if params.YearFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("n.release_year >= $%d", argIndex))
		args = append(args, *params.YearFrom)
		argIndex++
	}
	if params.YearTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("n.release_year <= $%d", argIndex))
		args = append(args, *params.YearTo)
		argIndex++
	}

	// Поиск
	if params.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf(`
			(nl.search_vector @@ plainto_tsquery('simple', $%d) OR nl.title ILIKE $%d)
		`, argIndex, argIndex+1))
		args = append(args, params.Search, "%"+params.Search+"%")
		argIndex += 2
	}

	// Собираем WHERE
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count novels: %w", err)
	}

	// Сортировка
	orderBy := "n.updated_at"
	switch params.Sort {
	case "created_at":
		orderBy = "n.created_at"
	case "views_daily":
		orderBy = "n.views_daily"
	case "views_total":
		orderBy = "n.views_total"
	case "rating":
		orderBy = "(n.rating_sum::float / NULLIF(n.rating_count, 0))"
	case "bookmarks_count":
		orderBy = "n.bookmarks_count"
	}

	order := "DESC"
	if params.Order == "asc" {
		order = "ASC"
	}

	// Main query
	selectQuery := fmt.Sprintf(`
		SELECT n.id, n.slug, n.cover_image_key, n.translation_status, n.original_chapters_count,
		       n.release_year, n.author, n.views_total, n.views_daily, n.rating_sum, n.rating_count,
		       n.bookmarks_count, n.created_at, n.updated_at,
		       nl.title, nl.description
		%s %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, order, argIndex, argIndex+1)

	offset := (params.Page - 1) * params.Limit
	args = append(args, params.Limit, offset)

	rows, err := r.db.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list novels: %w", err)
	}
	defer rows.Close()

	var novels []models.NovelWithLocalization
	for rows.Next() {
		var novel models.NovelWithLocalization
		err := rows.Scan(
			&novel.ID, &novel.Slug, &novel.CoverImageKey, &novel.TranslationStatus,
			&novel.OriginalChaptersCount, &novel.ReleaseYear, &novel.Author,
			&novel.ViewsTotal, &novel.ViewsDaily, &novel.RatingSum, &novel.RatingCount,
			&novel.BookmarksCount, &novel.CreatedAt, &novel.UpdatedAt,
			&novel.Title, &novel.Description,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan novel: %w", err)
		}

		// Рассчитываем рейтинг
		novel.Rating = novel.Novel.Rating()

		// Формируем URL обложки
		if novel.CoverImageKey != nil {
			url := "/uploads/" + *novel.CoverImageKey
			novel.CoverURL = &url
		}

		novels = append(novels, novel)
	}

	return novels, total, nil
}

// GetBySlug получает новеллу по slug
func (r *NovelRepository) GetBySlug(ctx context.Context, slug, lang string) (*models.NovelWithLocalization, error) {
	query := `
		SELECT n.id, n.slug, n.cover_image_key, n.translation_status, n.original_chapters_count,
		       n.release_year, n.author, n.views_total, n.views_daily, n.rating_sum, n.rating_count,
		       n.bookmarks_count, n.created_at, n.updated_at,
		       nl.title, nl.description, nl.alt_titles
		FROM novels n
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = $1
		WHERE n.slug = $2
	`

	var novel models.NovelWithLocalization
	var altTitles pq.StringArray
	err := r.db.QueryRowxContext(ctx, query, lang, slug).Scan(
		&novel.ID, &novel.Slug, &novel.CoverImageKey, &novel.TranslationStatus,
		&novel.OriginalChaptersCount, &novel.ReleaseYear, &novel.Author,
		&novel.ViewsTotal, &novel.ViewsDaily, &novel.RatingSum, &novel.RatingCount,
		&novel.BookmarksCount, &novel.CreatedAt, &novel.UpdatedAt,
		&novel.Title, &novel.Description, &altTitles,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}

	novel.AltTitles = altTitles
	novel.Rating = novel.Novel.Rating()

	if novel.CoverImageKey != nil {
		url := "/uploads/" + *novel.CoverImageKey
		novel.CoverURL = &url
	}

	// Получаем жанры
	genres, err := r.GetGenres(ctx, novel.ID, lang)
	if err != nil {
		return nil, err
	}
	novel.Genres = genres

	// Получаем теги
	tags, err := r.GetTags(ctx, novel.ID, lang)
	if err != nil {
		return nil, err
	}
	novel.Tags = tags

	return &novel, nil
}

// GetByID получает новеллу по ID
func (r *NovelRepository) GetByID(ctx context.Context, id uuid.UUID, lang string) (*models.NovelWithLocalization, error) {
	query := `
		SELECT n.id, n.slug, n.cover_image_key, n.translation_status, n.original_chapters_count,
		       n.release_year, n.author, n.views_total, n.views_daily, n.rating_sum, n.rating_count,
		       n.bookmarks_count, n.created_at, n.updated_at,
		       nl.title, nl.description, nl.alt_titles
		FROM novels n
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = $1
		WHERE n.id = $2
	`

	var novel models.NovelWithLocalization
	var altTitles pq.StringArray
	err := r.db.QueryRowxContext(ctx, query, lang, id).Scan(
		&novel.ID, &novel.Slug, &novel.CoverImageKey, &novel.TranslationStatus,
		&novel.OriginalChaptersCount, &novel.ReleaseYear, &novel.Author,
		&novel.ViewsTotal, &novel.ViewsDaily, &novel.RatingSum, &novel.RatingCount,
		&novel.BookmarksCount, &novel.CreatedAt, &novel.UpdatedAt,
		&novel.Title, &novel.Description, &altTitles,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}

	novel.AltTitles = altTitles
	novel.Rating = novel.Novel.Rating()

	return &novel, nil
}

// Create создает новую новеллу
func (r *NovelRepository) Create(ctx context.Context, req *models.CreateNovelRequest) (*models.Novel, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	novel := &models.Novel{
		ID:                    uuid.New(),
		Slug:                  req.Slug,
		TranslationStatus:     req.TranslationStatus,
		OriginalChaptersCount: req.OriginalChaptersCount,
		ReleaseYear:           req.ReleaseYear,
		Author:                req.Author,
	}

	// Вставляем новеллу
	query := `
		INSERT INTO novels (id, slug, translation_status, original_chapters_count, release_year, author)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	err = tx.QueryRowxContext(ctx, query, novel.ID, novel.Slug, novel.TranslationStatus,
		novel.OriginalChaptersCount, novel.ReleaseYear, novel.Author).Scan(&novel.CreatedAt, &novel.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create novel: %w", err)
	}

	// Вставляем локализации
	for _, loc := range req.Localizations {
		locQuery := `
			INSERT INTO novel_localizations (novel_id, lang, title, description, alt_titles)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.ExecContext(ctx, locQuery, novel.ID, loc.Lang, loc.Title, loc.Description, pq.Array(loc.AltTitles))
		if err != nil {
			return nil, fmt.Errorf("failed to create localization: %w", err)
		}
	}

	// Добавляем жанры
	for _, genreSlug := range req.Genres {
		genreQuery := `
			INSERT INTO novel_genres (novel_id, genre_id)
			SELECT $1, id FROM genres WHERE slug = $2
		`
		_, err = tx.ExecContext(ctx, genreQuery, novel.ID, genreSlug)
		if err != nil {
			return nil, fmt.Errorf("failed to add genre: %w", err)
		}
	}

	// Добавляем теги
	for _, tagSlug := range req.Tags {
		tagQuery := `
			INSERT INTO novel_tags (novel_id, tag_id)
			SELECT $1, id FROM tags WHERE slug = $2
		`
		_, err = tx.ExecContext(ctx, tagQuery, novel.ID, tagSlug)
		if err != nil {
			return nil, fmt.Errorf("failed to add tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return novel, nil
}

// Update обновляет новеллу
func (r *NovelRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateNovelRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Обновляем основную информацию
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Slug != nil {
		updates = append(updates, fmt.Sprintf("slug = $%d", argIndex))
		args = append(args, *req.Slug)
		argIndex++
	}
	if req.TranslationStatus != nil {
		updates = append(updates, fmt.Sprintf("translation_status = $%d", argIndex))
		args = append(args, *req.TranslationStatus)
		argIndex++
	}
	if req.OriginalChaptersCount != nil {
		updates = append(updates, fmt.Sprintf("original_chapters_count = $%d", argIndex))
		args = append(args, *req.OriginalChaptersCount)
		argIndex++
	}
	if req.ReleaseYear != nil {
		updates = append(updates, fmt.Sprintf("release_year = $%d", argIndex))
		args = append(args, *req.ReleaseYear)
		argIndex++
	}
	if req.Author != nil {
		updates = append(updates, fmt.Sprintf("author = $%d", argIndex))
		args = append(args, *req.Author)
		argIndex++
	}

	if len(updates) > 0 {
		query := fmt.Sprintf("UPDATE novels SET %s WHERE id = $%d", strings.Join(updates, ", "), argIndex)
		args = append(args, id)
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update novel: %w", err)
		}
	}

	// Обновляем локализации
	if len(req.Localizations) > 0 {
		for _, loc := range req.Localizations {
			locQuery := `
				INSERT INTO novel_localizations (novel_id, lang, title, description, alt_titles)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (novel_id, lang) DO UPDATE SET
					title = EXCLUDED.title,
					description = EXCLUDED.description,
					alt_titles = EXCLUDED.alt_titles
			`
			_, err = tx.ExecContext(ctx, locQuery, id, loc.Lang, loc.Title, loc.Description, pq.Array(loc.AltTitles))
			if err != nil {
				return fmt.Errorf("failed to update localization: %w", err)
			}
		}
	}

	// Обновляем жанры
	if req.Genres != nil {
		// Удаляем старые
		_, err = tx.ExecContext(ctx, "DELETE FROM novel_genres WHERE novel_id = $1", id)
		if err != nil {
			return fmt.Errorf("failed to delete genres: %w", err)
		}
		// Добавляем новые
		for _, genreSlug := range req.Genres {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO novel_genres (novel_id, genre_id)
				SELECT $1, id FROM genres WHERE slug = $2
			`, id, genreSlug)
			if err != nil {
				return fmt.Errorf("failed to add genre: %w", err)
			}
		}
	}

	// Обновляем теги
	if req.Tags != nil {
		// Удаляем старые
		_, err = tx.ExecContext(ctx, "DELETE FROM novel_tags WHERE novel_id = $1", id)
		if err != nil {
			return fmt.Errorf("failed to delete tags: %w", err)
		}
		// Добавляем новые
		for _, tagSlug := range req.Tags {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO novel_tags (novel_id, tag_id)
				SELECT $1, id FROM tags WHERE slug = $2
			`, id, tagSlug)
			if err != nil {
				return fmt.Errorf("failed to add tag: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Delete удаляет новеллу
func (r *NovelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM novels WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete novel: %w", err)
	}
	return nil
}

// GetGenres получает жанры новеллы
func (r *NovelRepository) GetGenres(ctx context.Context, novelID uuid.UUID, lang string) ([]models.Genre, error) {
	query := `
		SELECT g.id, g.slug, COALESCE(gl.name, g.slug) as name
		FROM genres g
		JOIN novel_genres ng ON g.id = ng.genre_id
		LEFT JOIN genre_localizations gl ON g.id = gl.genre_id AND gl.lang = $1
		WHERE ng.novel_id = $2
		ORDER BY gl.name
	`
	
	var genres []models.Genre
	err := r.db.SelectContext(ctx, &genres, query, lang, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres: %w", err)
	}
	return genres, nil
}

// GetTags получает теги новеллы
func (r *NovelRepository) GetTags(ctx context.Context, novelID uuid.UUID, lang string) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.slug, COALESCE(tl.name, t.slug) as name
		FROM tags t
		JOIN novel_tags nt ON t.id = nt.tag_id
		LEFT JOIN tag_localizations tl ON t.id = tl.tag_id AND tl.lang = $1
		WHERE nt.novel_id = $2
		ORDER BY tl.name
	`
	
	var tags []models.Tag
	err := r.db.SelectContext(ctx, &tags, query, lang, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	return tags, nil
}

// ListAllGenres получает все жанры
func (r *NovelRepository) ListAllGenres(ctx context.Context, lang string) ([]models.Genre, error) {
	query := `
		SELECT g.id, g.slug, COALESCE(gl.name, g.slug) as name
		FROM genres g
		LEFT JOIN genre_localizations gl ON g.id = gl.genre_id AND gl.lang = $1
		ORDER BY gl.name
	`
	
	var genres []models.Genre
	err := r.db.SelectContext(ctx, &genres, query, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to list genres: %w", err)
	}
	return genres, nil
}

// ListAllTags получает все теги
func (r *NovelRepository) ListAllTags(ctx context.Context, lang string) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.slug, COALESCE(tl.name, t.slug) as name
		FROM tags t
		LEFT JOIN tag_localizations tl ON t.id = tl.tag_id AND tl.lang = $1
		ORDER BY tl.name
	`
	
	var tags []models.Tag
	err := r.db.SelectContext(ctx, &tags, query, lang)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return tags, nil
}

// IncrementViews увеличивает счетчик просмотров
func (r *NovelRepository) IncrementViews(ctx context.Context, novelID uuid.UUID) error {
	query := `UPDATE novels SET views_total = views_total + 1, views_daily = views_daily + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, novelID)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}
	return nil
}

// UpdateCoverImage обновляет обложку новеллы
func (r *NovelRepository) UpdateCoverImage(ctx context.Context, novelID uuid.UUID, imageKey string) error {
	query := `UPDATE novels SET cover_image_key = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, imageKey, novelID)
	if err != nil {
		return fmt.Errorf("failed to update cover image: %w", err)
	}
	return nil
}
