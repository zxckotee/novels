package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TranslationStatus представляет статус перевода
type TranslationStatus string

const (
	StatusOngoing   TranslationStatus = "ongoing"
	StatusCompleted TranslationStatus = "completed"
	StatusPaused    TranslationStatus = "paused"
	StatusDropped   TranslationStatus = "dropped"
)

// Novel представляет новеллу
type Novel struct {
	ID                    uuid.UUID         `db:"id" json:"id"`
	Slug                  string            `db:"slug" json:"slug"`
	CoverImageKey         *string           `db:"cover_image_key" json:"cover_image_key,omitempty"`
	TranslationStatus     TranslationStatus `db:"translation_status" json:"translation_status"`
	OriginalChaptersCount int               `db:"original_chapters_count" json:"original_chapters_count"`
	ReleaseYear           *int              `db:"release_year" json:"release_year,omitempty"`
	Author                *string           `db:"author" json:"author,omitempty"`
	ViewsTotal            int64             `db:"views_total" json:"views_total"`
	ViewsDaily            int               `db:"views_daily" json:"views_daily"`
	RatingSum             int               `db:"rating_sum" json:"-"`
	RatingCount           int               `db:"rating_count" json:"rating_count"`
	BookmarksCount        int               `db:"bookmarks_count" json:"bookmarks_count"`
	CreatedAt             time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time         `db:"updated_at" json:"updated_at"`
}

// Rating возвращает средний рейтинг
func (n *Novel) Rating() float64 {
	if n.RatingCount == 0 {
		return 0
	}
	return float64(n.RatingSum) / float64(n.RatingCount)
}

// NovelLocalization представляет локализацию новеллы
type NovelLocalization struct {
	NovelID     uuid.UUID      `db:"novel_id" json:"novel_id"`
	Lang        string         `db:"lang" json:"lang"`
	Title       string         `db:"title" json:"title"`
	Description *string        `db:"description" json:"description,omitempty"`
	AltTitles   pq.StringArray `db:"alt_titles" json:"alt_titles,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// NovelWithLocalization объединяет новеллу с её локализацией
type NovelWithLocalization struct {
	Novel
	Title       string   `db:"title" json:"title"`
	Description *string  `db:"description" json:"description,omitempty"`
	AltTitles   []string `json:"alt_titles,omitempty"`
	CoverURL    *string  `json:"cover_url,omitempty"`
	Rating      float64  `json:"rating"`
	Genres      []Genre  `json:"genres,omitempty"`
	Tags        []Tag    `json:"tags,omitempty"`
}

// NovelCard представляет облегчённую карточку новеллы для списков.
// Используется как псевдоним NovelWithLocalization.
type NovelCard = NovelWithLocalization

// NovelDetail представляет детальную информацию о новелле
type NovelDetail struct {
	NovelWithLocalization
	ChaptersPublished int            `json:"chapters_published"`
	LastChapterAt     *time.Time     `json:"last_chapter_at,omitempty"`
	UserProgress      *UserProgress  `json:"user_progress,omitempty"`
	UserBookmark      *UserBookmark  `json:"user_bookmark,omitempty"`
	UserRating        *int           `json:"user_rating,omitempty"`
}

// UserProgress представляет прогресс пользователя
type UserProgress struct {
	ChapterID     uuid.UUID `json:"chapter_id"`
	ChapterNumber float64   `json:"chapter_number"`
	Position      int       `json:"position"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserBookmark представляет закладку пользователя
type UserBookmark struct {
	ListCode  string    `json:"list_code"`
	ListTitle string    `json:"list_title"`
	CreatedAt time.Time `json:"created_at"`
}

// Genre представляет жанр
type Genre struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Slug string    `db:"slug" json:"slug"`
	Name string    `db:"name" json:"name"`
}

// Tag представляет тег
type Tag struct {
	ID   uuid.UUID `db:"id" json:"id"`
	Slug string    `db:"slug" json:"slug"`
	Name string    `db:"name" json:"name"`
}

// NovelListParams параметры для списка новелл
type NovelListParams struct {
	Lang        string   `json:"lang"`
	Page        int      `json:"page"`
	Limit       int      `json:"limit"`
	Sort        string   `json:"sort"`
	Order       string   `json:"order"`
	Status      []string `json:"status,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Search      string   `json:"search,omitempty"`
	YearFrom    *int     `json:"year_from,omitempty"`
	YearTo      *int     `json:"year_to,omitempty"`
}

// CreateNovelRequest запрос на создание новеллы
type CreateNovelRequest struct {
	Slug                  string                      `json:"slug" validate:"required,min=1,max=255"`
	TranslationStatus     TranslationStatus           `json:"translation_status" validate:"required,oneof=ongoing completed paused dropped"`
	OriginalChaptersCount int                         `json:"original_chapters_count" validate:"gte=0"`
	ReleaseYear           *int                        `json:"release_year" validate:"omitempty,gte=1900,lte=2100"`
	Author                *string                     `json:"author" validate:"omitempty,max=255"`
	CoverImage            *string                     `json:"cover_image,omitempty"` // base64 encoded
	Localizations         []CreateLocalizationRequest `json:"localizations" validate:"required,min=1,dive"`
	Genres                []string                    `json:"genres,omitempty"`
	Tags                  []string                    `json:"tags,omitempty"`
}

// CreateLocalizationRequest запрос на создание локализации
type CreateLocalizationRequest struct {
	Lang        string   `json:"lang" validate:"required,len=2"`
	Title       string   `json:"title" validate:"required,min=1,max=500"`
	Description *string  `json:"description,omitempty"`
	AltTitles   []string `json:"alt_titles,omitempty"`
}

// UpdateNovelRequest запрос на обновление новеллы
type UpdateNovelRequest struct {
	Slug                  *string                     `json:"slug,omitempty" validate:"omitempty,min=1,max=255"`
	TranslationStatus     *TranslationStatus          `json:"translation_status,omitempty" validate:"omitempty,oneof=ongoing completed paused dropped"`
	OriginalChaptersCount *int                        `json:"original_chapters_count,omitempty" validate:"omitempty,gte=0"`
	ReleaseYear           *int                        `json:"release_year,omitempty" validate:"omitempty,gte=1900,lte=2100"`
	Author                *string                     `json:"author,omitempty" validate:"omitempty,max=255"`
	CoverImage            *string                     `json:"cover_image,omitempty"`
	Localizations         []CreateLocalizationRequest `json:"localizations,omitempty" validate:"omitempty,dive"`
	Genres                []string                    `json:"genres,omitempty"`
	Tags                  []string                    `json:"tags,omitempty"`
}

// NovelListResponse ответ со списком новелл
type NovelListResponse struct {
	Novels     []NovelWithLocalization `json:"novels"`
	Filters    *NovelFilters           `json:"filters,omitempty"`
	Pagination Pagination              `json:"pagination"`
}

// Pagination описывает простую пагинацию для списков новелл
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// NovelFilters фильтры для каталога
type NovelFilters struct {
	Genres []Genre `json:"genres"`
	Tags   []Tag   `json:"tags"`
	Years  []int   `json:"years"`
}
