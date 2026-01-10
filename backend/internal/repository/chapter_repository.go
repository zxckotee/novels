package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ChapterRepository репозиторий для работы с главами
type ChapterRepository struct {
	db *sqlx.DB
}

// NewChapterRepository создает новый ChapterRepository
func NewChapterRepository(db *sqlx.DB) *ChapterRepository {
	return &ChapterRepository{db: db}
}

// ListByNovel получает список глав новеллы
func (r *ChapterRepository) ListByNovel(ctx context.Context, novelSlug string, params models.ChapterListParams) ([]models.ChapterListItem, *models.NovelBrief, int, error) {
	// Получаем ID новеллы и краткую информацию
	var novel models.NovelBrief
	novelQuery := `
		SELECT n.id, n.slug, nl.title, n.cover_image_key
		FROM novels n
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = 'ru'
		WHERE n.slug = $1
	`
	var coverKey *string
	err := r.db.QueryRowxContext(ctx, novelQuery, novelSlug).Scan(&novel.ID, &novel.Slug, &novel.Title, &coverKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, 0, nil
		}
		return nil, nil, 0, fmt.Errorf("failed to get novel: %w", err)
	}

	if coverKey != nil {
		url := "/uploads/" + *coverKey
		novel.CoverURL = &url
	}

	// Подсчет глав
	countQuery := `SELECT COUNT(*) FROM chapters WHERE novel_id = $1`
	var total int
	err = r.db.GetContext(ctx, &total, countQuery, novel.ID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to count chapters: %w", err)
	}

	// Сортировка
	orderBy := "c.number"
	switch params.Sort {
	case "created_at":
		orderBy = "c.created_at"
	case "views":
		orderBy = "c.views"
	}

	order := "ASC"
	if params.Order == "desc" {
		order = "DESC"
	}

	// Получаем главы
	offset := (params.Page - 1) * params.Limit
	chaptersQuery := fmt.Sprintf(`
		SELECT c.id, c.number, c.slug, c.title, c.views, c.published_at
		FROM chapters c
		WHERE c.novel_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, orderBy, order)

	rows, err := r.db.QueryxContext(ctx, chaptersQuery, novel.ID, params.Limit, offset)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to list chapters: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	twentyFourHoursAgo := now.Add(-24 * time.Hour)

	var chapters []models.ChapterListItem
	for rows.Next() {
		var chapter models.ChapterListItem
		err := rows.Scan(&chapter.ID, &chapter.Number, &chapter.Slug, &chapter.Title, &chapter.Views, &chapter.PublishedAt)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to scan chapter: %w", err)
		}

		// Проверяем, новая ли глава
		if chapter.PublishedAt != nil && chapter.PublishedAt.After(twentyFourHoursAgo) {
			chapter.IsNew = true
		}

		chapters = append(chapters, chapter)
	}

	return chapters, &novel, total, nil
}

// GetByID получает главу по ID с содержимым
func (r *ChapterRepository) GetByID(ctx context.Context, id uuid.UUID, lang string) (*models.ChapterWithContent, error) {
	query := `
		SELECT c.id, c.novel_id, c.number, c.slug, c.title, c.views, c.published_at, c.created_at, c.updated_at,
		       cc.content, cc.word_count, cc.source,
		       n.slug as novel_slug, nl.title as novel_title
		FROM chapters c
		JOIN chapter_contents cc ON c.id = cc.chapter_id AND cc.lang = $1
		JOIN novels n ON c.novel_id = n.id
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = $1
		WHERE c.id = $2
	`

	var chapter models.ChapterWithContent
	err := r.db.QueryRowxContext(ctx, query, lang, id).Scan(
		&chapter.ID, &chapter.NovelID, &chapter.Number, &chapter.Slug, &chapter.Title,
		&chapter.Views, &chapter.PublishedAt, &chapter.CreatedAt, &chapter.UpdatedAt,
		&chapter.Content, &chapter.WordCount, &chapter.Source,
		&chapter.NovelSlug, &chapter.NovelTitle,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	// Примерное время чтения (200 слов в минуту)
	chapter.ReadingTime = chapter.WordCount / 200
	if chapter.ReadingTime < 1 {
		chapter.ReadingTime = 1
	}

	// Получаем соседние главы
	prevQuery := `
		SELECT id, number, title FROM chapters
		WHERE novel_id = $1 AND number < $2
		ORDER BY number DESC LIMIT 1
	`
	var prevChapter models.ChapterNavInfo
	err = r.db.QueryRowxContext(ctx, prevQuery, chapter.NovelID, chapter.Number).Scan(
		&prevChapter.ID, &prevChapter.Number, &prevChapter.Title,
	)
	if err == nil {
		chapter.PrevChapter = &prevChapter
	}

	nextQuery := `
		SELECT id, number, title FROM chapters
		WHERE novel_id = $1 AND number > $2
		ORDER BY number ASC LIMIT 1
	`
	var nextChapter models.ChapterNavInfo
	err = r.db.QueryRowxContext(ctx, nextQuery, chapter.NovelID, chapter.Number).Scan(
		&nextChapter.ID, &nextChapter.Number, &nextChapter.Title,
	)
	if err == nil {
		chapter.NextChapter = &nextChapter
	}

	return &chapter, nil
}

// Create создает новую главу
func (r *ChapterRepository) Create(ctx context.Context, req *models.CreateChapterRequest) (*models.Chapter, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	chapter := &models.Chapter{
		ID:          uuid.New(),
		NovelID:     req.NovelID,
		Number:      req.Number,
		Slug:        req.Slug,
		Title:       req.Title,
		PublishedAt: timePtr(time.Now()),
	}

	// Вставляем главу
	query := `
		INSERT INTO chapters (id, novel_id, number, slug, title, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	err = tx.QueryRowxContext(ctx, query, chapter.ID, chapter.NovelID, chapter.Number,
		chapter.Slug, chapter.Title, chapter.PublishedAt).Scan(&chapter.CreatedAt, &chapter.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create chapter: %w", err)
	}

	// Вставляем содержимое
	for _, content := range req.Contents {
		source := content.Source
		if source == "" {
			source = "manual"
		}
		
		contentQuery := `
			INSERT INTO chapter_contents (chapter_id, lang, content, source)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.ExecContext(ctx, contentQuery, chapter.ID, content.Lang, content.Content, source)
		if err != nil {
			return nil, fmt.Errorf("failed to create chapter content: %w", err)
		}
	}

	// Обновляем updated_at новеллы
	_, err = tx.ExecContext(ctx, "UPDATE novels SET updated_at = NOW() WHERE id = $1", req.NovelID)
	if err != nil {
		return nil, fmt.Errorf("failed to update novel: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return chapter, nil
}

// Update обновляет главу
func (r *ChapterRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateChapterRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Обновляем основную информацию
	if req.Number != nil || req.Slug != nil || req.Title != nil {
		updates := []string{}
		args := []interface{}{}
		argIndex := 1

		if req.Number != nil {
			updates = append(updates, fmt.Sprintf("number = $%d", argIndex))
			args = append(args, *req.Number)
			argIndex++
		}
		if req.Slug != nil {
			updates = append(updates, fmt.Sprintf("slug = $%d", argIndex))
			args = append(args, *req.Slug)
			argIndex++
		}
		if req.Title != nil {
			updates = append(updates, fmt.Sprintf("title = $%d", argIndex))
			args = append(args, *req.Title)
			argIndex++
		}

		query := fmt.Sprintf("UPDATE chapters SET %s WHERE id = $%d", 
			joinStrings(updates, ", "), argIndex)
		args = append(args, id)
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update chapter: %w", err)
		}
	}

	// Обновляем содержимое
	if len(req.Contents) > 0 {
		for _, content := range req.Contents {
			source := content.Source
			if source == "" {
				source = "manual"
			}

			contentQuery := `
				INSERT INTO chapter_contents (chapter_id, lang, content, source)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (chapter_id, lang) DO UPDATE SET
					content = EXCLUDED.content,
					source = EXCLUDED.source
			`
			_, err = tx.ExecContext(ctx, contentQuery, id, content.Lang, content.Content, source)
			if err != nil {
				return fmt.Errorf("failed to update chapter content: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Delete удаляет главу
func (r *ChapterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM chapters WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}
	return nil
}

// IncrementViews увеличивает счетчик просмотров
func (r *ChapterRepository) IncrementViews(ctx context.Context, chapterID uuid.UUID) error {
	query := `UPDATE chapters SET views = views + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, chapterID)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}
	return nil
}

// GetNovelIDByChapter получает ID новеллы по ID главы
func (r *ChapterRepository) GetNovelIDByChapter(ctx context.Context, chapterID uuid.UUID) (uuid.UUID, error) {
	var novelID uuid.UUID
	query := `SELECT novel_id FROM chapters WHERE id = $1`
	err := r.db.GetContext(ctx, &novelID, query, chapterID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to get novel id: %w", err)
	}
	return novelID, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func joinStrings(s []string, sep string) string {
	result := ""
	for i, str := range s {
		if i > 0 {
			result += sep
		}
		result += str
	}
	return result
}
