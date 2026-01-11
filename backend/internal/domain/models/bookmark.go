package models

import (
	"time"

	"github.com/google/uuid"
)

// BookmarkListCode represents system bookmark list types
type BookmarkListCode string

const (
	BookmarkListReading   BookmarkListCode = "reading"
	BookmarkListPlanned   BookmarkListCode = "planned"
	BookmarkListDropped   BookmarkListCode = "dropped"
	BookmarkListCompleted BookmarkListCode = "completed"
	BookmarkListFavorites BookmarkListCode = "favorites"
)

var SystemBookmarkLists = []BookmarkListCode{
	BookmarkListReading,
	BookmarkListPlanned,
	BookmarkListDropped,
	BookmarkListCompleted,
	BookmarkListFavorites,
}

// BookmarkList represents a bookmark list (system or custom)
type BookmarkList struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"userId" db:"user_id"`
	Code      BookmarkListCode `json:"code" db:"code"`
	Title     string           `json:"title" db:"title"`
	SortOrder int              `json:"sortOrder" db:"sort_order"`
	IsSystem  bool             `json:"isSystem" db:"is_system"`
	Count     int              `json:"count" db:"count"` // Computed field
	CreatedAt time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time        `json:"updatedAt" db:"updated_at"`
}

// Bookmark represents a novel in a bookmark list
type Bookmark struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	NovelID   uuid.UUID `json:"novelId" db:"novel_id"`
	ListID    uuid.UUID `json:"listId" db:"list_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	
	// Populated from joins
	Novel         *BookmarkNovel          `json:"novel,omitempty"`
	List          *BookmarkList           `json:"list,omitempty"`
	Progress      *BookmarkReadingProgress `json:"progress,omitempty"`
	LatestChapter *ChapterInfo            `json:"latestChapter,omitempty"`
	HasNewChapter bool                    `json:"hasNewChapter"`
}

// BookmarkNovel represents minimal novel info for bookmarks
type BookmarkNovel struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	Slug              string            `json:"slug" db:"slug"`
	CoverImageKey     *string           `json:"coverImageKey,omitempty" db:"cover_image_key"`
	TranslationStatus TranslationStatus `json:"translationStatus" db:"translation_status"`
	Title             string            `json:"title" db:"title"`
	ChaptersCount     int               `json:"chaptersCount" db:"chapters_count"`
	Rating            float64           `json:"rating" db:"rating"`
}

// ChapterInfo represents minimal chapter info
type ChapterInfo struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Number      int       `json:"number" db:"number"`
	Title       string    `json:"title" db:"title"`
	PublishedAt time.Time `json:"publishedAt" db:"published_at"`
}

// BookmarkReadingProgress represents reading progress info used in bookmarks context
type BookmarkReadingProgress struct {
	ChapterID   uuid.UUID `json:"chapterId" db:"chapter_id"`
	ChapterNum  int       `json:"chapterNum" db:"chapter_num"`
	TotalChapters int     `json:"totalChapters" db:"total_chapters"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// CreateBookmarkRequest represents request to add novel to bookmark list
type CreateBookmarkRequest struct {
	NovelID string `json:"novelId" validate:"required,uuid"`
	ListCode BookmarkListCode `json:"listCode" validate:"required"`
}

// UpdateBookmarkRequest represents request to move bookmark to another list
type UpdateBookmarkRequest struct {
	ListCode BookmarkListCode `json:"listCode" validate:"required"`
}

// BookmarksFilter represents filters for listing bookmarks
type BookmarksFilter struct {
	UserID   uuid.UUID
	ListCode *BookmarkListCode
	Sort     string // latest_update, date_added, title
	Page     int
	Limit    int
}

// BookmarksResponse represents paginated bookmarks response
type BookmarksResponse struct {
	Bookmarks  []Bookmark     `json:"bookmarks"`
	Lists      []BookmarkList `json:"lists"`
	TotalCount int            `json:"totalCount"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// BookmarkListStats represents stats for a bookmark list
type BookmarkListStats struct {
	ListCode BookmarkListCode `json:"listCode"`
	Count    int              `json:"count"`
}

// GetListTitle returns localized title for system list
func GetListTitle(code BookmarkListCode, lang string) string {
	titles := map[BookmarkListCode]map[string]string{
		BookmarkListReading: {
			"ru": "Читаю",
			"en": "Reading",
		},
		BookmarkListPlanned: {
			"ru": "В планах",
			"en": "Plan to Read",
		},
		BookmarkListDropped: {
			"ru": "Брошено",
			"en": "Dropped",
		},
		BookmarkListCompleted: {
			"ru": "Прочитано",
			"en": "Completed",
		},
		BookmarkListFavorites: {
			"ru": "Любимые",
			"en": "Favorites",
		},
	}
	
	if codeMap, ok := titles[code]; ok {
		if title, ok := codeMap[lang]; ok {
			return title
		}
		return codeMap["en"]
	}
	return string(code)
}
