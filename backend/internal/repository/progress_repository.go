package repository

import (
	"context"
	"database/sql"
	"fmt"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ProgressRepository репозиторий для работы с прогрессом чтения
type ProgressRepository struct {
	db *sqlx.DB
}

// NewProgressRepository создает новый ProgressRepository
func NewProgressRepository(db *sqlx.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// Get получает прогресс чтения пользователя для новеллы
func (r *ProgressRepository) Get(ctx context.Context, userID, novelID uuid.UUID) (*models.ReadingProgressWithChapter, error) {
	query := `
		SELECT rp.user_id, rp.novel_id, rp.chapter_id, rp.position, rp.updated_at,
		       c.number as chapter_number, c.title as chapter_title
		FROM reading_progress rp
		JOIN chapters c ON rp.chapter_id = c.id
		WHERE rp.user_id = $1 AND rp.novel_id = $2
	`

	var progress models.ReadingProgressWithChapter
	err := r.db.GetContext(ctx, &progress, query, userID, novelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get reading progress: %w", err)
	}

	return &progress, nil
}

// GetAll получает весь прогресс чтения пользователя
func (r *ProgressRepository) GetAll(ctx context.Context, userID uuid.UUID) ([]models.ReadingProgressWithChapter, error) {
	query := `
		SELECT rp.user_id, rp.novel_id, rp.chapter_id, rp.position, rp.updated_at,
		       c.number as chapter_number, c.title as chapter_title
		FROM reading_progress rp
		JOIN chapters c ON rp.chapter_id = c.id
		WHERE rp.user_id = $1
		ORDER BY rp.updated_at DESC
	`

	var progress []models.ReadingProgressWithChapter
	err := r.db.SelectContext(ctx, &progress, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reading progress: %w", err)
	}

	return progress, nil
}

// Save сохраняет или обновляет прогресс чтения
func (r *ProgressRepository) Save(ctx context.Context, progress *models.ReadingProgress) error {
	query := `
		INSERT INTO reading_progress (user_id, novel_id, chapter_id, position, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, novel_id) DO UPDATE SET
			chapter_id = EXCLUDED.chapter_id,
			position = EXCLUDED.position,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query, progress.UserID, progress.NovelID, progress.ChapterID, progress.Position)
	if err != nil {
		return fmt.Errorf("failed to save reading progress: %w", err)
	}

	return nil
}

// Delete удаляет прогресс чтения
func (r *ProgressRepository) Delete(ctx context.Context, userID, novelID uuid.UUID) error {
	query := `DELETE FROM reading_progress WHERE user_id = $1 AND novel_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, novelID)
	if err != nil {
		return fmt.Errorf("failed to delete reading progress: %w", err)
	}
	return nil
}

// GetRecentlyRead получает недавно прочитанные новеллы
func (r *ProgressRepository) GetRecentlyRead(ctx context.Context, userID uuid.UUID, limit int) ([]models.NovelWithLocalization, error) {
	query := `
		SELECT n.id, n.slug, n.cover_image_key, n.translation_status, n.original_chapters_count,
		       n.release_year, n.author, n.views_total, n.views_daily, n.rating_sum, n.rating_count,
		       n.bookmarks_count, n.created_at, n.updated_at,
		       nl.title, nl.description
		FROM reading_progress rp
		JOIN novels n ON rp.novel_id = n.id
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = 'ru'
		WHERE rp.user_id = $1
		ORDER BY rp.updated_at DESC
		LIMIT $2
	`

	var novels []models.NovelWithLocalization
	rows, err := r.db.QueryxContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently read: %w", err)
	}
	defer rows.Close()

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
			return nil, fmt.Errorf("failed to scan novel: %w", err)
		}

		novel.Rating = novel.Novel.Rating()
		if novel.CoverImageKey != nil {
			url := "/uploads/" + *novel.CoverImageKey
			novel.CoverURL = &url
		}

		novels = append(novels, novel)
	}

	return novels, nil
}

// MarkChaptersAsRead отмечает главы как прочитанные для пользователя
func (r *ProgressRepository) MarkChaptersAsRead(ctx context.Context, userID uuid.UUID, chapterIDs []uuid.UUID) error {
	if len(chapterIDs) == 0 {
		return nil
	}

	// Для каждой главы обновляем прогресс
	for _, chapterID := range chapterIDs {
		// Получаем novel_id для главы
		var novelID uuid.UUID
		err := r.db.GetContext(ctx, &novelID, "SELECT novel_id FROM chapters WHERE id = $1", chapterID)
		if err != nil {
			continue // Пропускаем если глава не найдена
		}

		progress := &models.ReadingProgress{
			UserID:    userID,
			NovelID:   novelID,
			ChapterID: chapterID,
			Position:  0,
		}

		if err := r.Save(ctx, progress); err != nil {
			return err
		}
	}

	return nil
}

// GetReadChapterIDs получает ID прочитанных глав для новеллы
func (r *ProgressRepository) GetReadChapterIDs(ctx context.Context, userID, novelID uuid.UUID) ([]uuid.UUID, error) {
	// В текущей реализации у нас только одна последняя глава в прогрессе
	// Для полноценного отслеживания нужна отдельная таблица
	// Пока возвращаем ID текущей главы
	
	progress, err := r.Get(ctx, userID, novelID)
	if err != nil {
		return nil, err
	}

	if progress == nil {
		return []uuid.UUID{}, nil
	}

	return []uuid.UUID{progress.ChapterID}, nil
}
