package models

import (
	"time"

	"github.com/google/uuid"
)

// Chapter представляет главу новеллы
type Chapter struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	NovelID     uuid.UUID  `db:"novel_id" json:"novel_id"`
	Number      float64    `db:"number" json:"number"` // Поддержка дробных номеров (1.5, 2.1)
	Slug        *string    `db:"slug" json:"slug,omitempty"`
	Title       *string    `db:"title" json:"title,omitempty"`
	Views       int        `db:"views" json:"views"`
	PublishedAt *time.Time `db:"published_at" json:"published_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// ChapterContent представляет содержимое главы
type ChapterContent struct {
	ChapterID uuid.UUID `db:"chapter_id" json:"chapter_id"`
	Lang      string    `db:"lang" json:"lang"`
	Content   string    `db:"content" json:"content"`
	WordCount int       `db:"word_count" json:"word_count"`
	Source    string    `db:"source" json:"source"` // manual, auto, import
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ChapterWithContent объединяет главу с её содержимым
type ChapterWithContent struct {
	Chapter
	Content          string          `json:"content"`
	WordCount        int             `json:"word_count"`
	ReadingTime      int             `json:"reading_time_minutes"` // Примерное время чтения
	Source           string          `json:"source"`
	PrevChapter      *ChapterNavInfo `json:"prev_chapter,omitempty"`
	NextChapter      *ChapterNavInfo `json:"next_chapter,omitempty"`
	NovelSlug        string          `json:"novel_slug"`
	NovelTitle       string          `json:"novel_title"`
}

// ChapterNavInfo информация для навигации между главами
type ChapterNavInfo struct {
	ID     uuid.UUID `json:"id"`
	Number float64   `json:"number"`
	Title  *string   `json:"title,omitempty"`
}

// ChapterListItem элемент списка глав
type ChapterListItem struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	Number        float64    `db:"number" json:"number"`
	Slug          *string    `db:"slug" json:"slug,omitempty"`
	Title         *string    `db:"title" json:"title,omitempty"`
	Views         int        `db:"views" json:"views"`
	PublishedAt   *time.Time `db:"published_at" json:"published_at,omitempty"`
	IsRead        bool       `json:"is_read"`
	IsNew         bool       `json:"is_new"` // Вышла за последние 24 часа
	CommentsCount int        `json:"comments_count"`
}

// ChapterListParams параметры для списка глав
type ChapterListParams struct {
	NovelSlug string `json:"novel_slug"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Sort      string `json:"sort"`  // number, created_at, views
	Order     string `json:"order"` // asc, desc
}

// CreateChapterRequest запрос на создание главы
type CreateChapterRequest struct {
	NovelID  uuid.UUID                     `json:"novel_id" validate:"required"`
	Number   float64                       `json:"number" validate:"required,gt=0"`
	Slug     *string                       `json:"slug,omitempty" validate:"omitempty,max=255"`
	Title    *string                       `json:"title,omitempty" validate:"omitempty,max=500"`
	Contents []CreateChapterContentRequest `json:"contents" validate:"required,min=1,dive"`
}

// CreateChapterContentRequest запрос на создание содержимого главы
type CreateChapterContentRequest struct {
	Lang    string `json:"lang" validate:"required,len=2"`
	Content string `json:"content" validate:"required"`
	Source  string `json:"source,omitempty" validate:"omitempty,oneof=manual auto import"`
}

// UpdateChapterRequest запрос на обновление главы
type UpdateChapterRequest struct {
	Number   *float64                      `json:"number,omitempty" validate:"omitempty,gt=0"`
	Slug     *string                       `json:"slug,omitempty" validate:"omitempty,max=255"`
	Title    *string                       `json:"title,omitempty" validate:"omitempty,max=500"`
	Contents []CreateChapterContentRequest `json:"contents,omitempty" validate:"omitempty,dive"`
}

// ReadingProgress прогресс чтения
type ReadingProgress struct {
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	NovelID   uuid.UUID `db:"novel_id" json:"novel_id"`
	ChapterID uuid.UUID `db:"chapter_id" json:"chapter_id"`
	Position  int       `db:"position" json:"position"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ReadingProgressWithChapter прогресс чтения с информацией о главе
type ReadingProgressWithChapter struct {
	ReadingProgress
	ChapterNumber float64 `db:"chapter_number" json:"chapter_number"`
	ChapterTitle  *string `db:"chapter_title" json:"chapter_title,omitempty"`
}

// SaveProgressRequest запрос на сохранение прогресса
type SaveProgressRequest struct {
	NovelID   uuid.UUID `json:"novel_id" validate:"required"`
	ChapterID uuid.UUID `json:"chapter_id" validate:"required"`
	Position  int       `json:"position" validate:"gte=0"`
}

// ChaptersListResponse ответ со списком глав
type ChaptersListResponse struct {
	Chapters []ChapterListItem `json:"chapters"`
	Novel    *NovelBrief       `json:"novel"`
	Pagination Pagination      `json:"pagination"`
}

// NovelBrief краткая информация о новелле
type NovelBrief struct {
	ID       uuid.UUID `json:"id"`
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	CoverURL *string   `json:"cover_url,omitempty"`
}
