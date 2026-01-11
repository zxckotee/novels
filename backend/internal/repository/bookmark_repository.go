package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
)

type BookmarkRepository struct {
	db *sqlx.DB
}

func NewBookmarkRepository(db *sqlx.DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

// GetOrCreateLists gets or creates system bookmark lists for a user
func (r *BookmarkRepository) GetOrCreateLists(ctx context.Context, userID uuid.UUID) ([]models.BookmarkList, error) {
	// First, try to get existing lists
	var lists []models.BookmarkList
	query := `
		SELECT id, user_id, code, title, sort_order, is_system, created_at, updated_at
		FROM bookmark_lists
		WHERE user_id = $1
		ORDER BY sort_order`
	
	err := r.db.SelectContext(ctx, &lists, query, userID)
	if err != nil {
		return nil, err
	}
	
	// If lists exist, return them with counts
	if len(lists) > 0 {
		return r.getListsWithCounts(ctx, userID, lists)
	}
	
	// Create system lists
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	
	for i, code := range models.SystemBookmarkLists {
		list := models.BookmarkList{
			ID:        uuid.New(),
			UserID:    userID,
			Code:      code,
			Title:     models.GetListTitle(code, "ru"),
			SortOrder: i,
			IsSystem:  true,
		}
		
		_, err = tx.ExecContext(ctx, `
			INSERT INTO bookmark_lists (id, user_id, code, title, sort_order, is_system, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
			list.ID, list.UserID, list.Code, list.Title, list.SortOrder, list.IsSystem)
		if err != nil {
			return nil, err
		}
		
		lists = append(lists, list)
	}
	
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	
	return lists, nil
}

// getListsWithCounts adds bookmark counts to lists
func (r *BookmarkRepository) getListsWithCounts(ctx context.Context, userID uuid.UUID, lists []models.BookmarkList) ([]models.BookmarkList, error) {
	query := `
		SELECT list_id, COUNT(*) as count
		FROM bookmarks
		WHERE user_id = $1
		GROUP BY list_id`
	
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var listID uuid.UUID
		var count int
		if err := rows.Scan(&listID, &count); err != nil {
			return nil, err
		}
		counts[listID] = count
	}
	
	for i := range lists {
		lists[i].Count = counts[lists[i].ID]
	}
	
	return lists, nil
}

// GetListByCode gets a bookmark list by code
func (r *BookmarkRepository) GetListByCode(ctx context.Context, userID uuid.UUID, code models.BookmarkListCode) (*models.BookmarkList, error) {
	var list models.BookmarkList
	query := `
		SELECT id, user_id, code, title, sort_order, is_system, created_at, updated_at
		FROM bookmark_lists
		WHERE user_id = $1 AND code = $2`
	
	err := r.db.GetContext(ctx, &list, query, userID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &list, nil
}

// GetBookmark gets a bookmark for a novel
func (r *BookmarkRepository) GetBookmark(ctx context.Context, userID, novelID uuid.UUID) (*models.Bookmark, error) {
	var bookmark models.Bookmark
	query := `
		SELECT id, user_id, novel_id, list_id, created_at, updated_at
		FROM bookmarks
		WHERE user_id = $1 AND novel_id = $2`
	
	err := r.db.GetContext(ctx, &bookmark, query, userID, novelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &bookmark, nil
}

// Create creates a new bookmark
func (r *BookmarkRepository) Create(ctx context.Context, bookmark *models.Bookmark) error {
	query := `
		INSERT INTO bookmarks (id, user_id, novel_id, list_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id, novel_id) DO UPDATE SET
			list_id = $4,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`
	
	return r.db.QueryRowxContext(ctx, query,
		bookmark.ID,
		bookmark.UserID,
		bookmark.NovelID,
		bookmark.ListID,
	).Scan(&bookmark.ID, &bookmark.CreatedAt, &bookmark.UpdatedAt)
}

// Update updates bookmark's list
func (r *BookmarkRepository) Update(ctx context.Context, userID, novelID, listID uuid.UUID) error {
	query := `
		UPDATE bookmarks
		SET list_id = $3, updated_at = NOW()
		WHERE user_id = $1 AND novel_id = $2`
	
	result, err := r.db.ExecContext(ctx, query, userID, novelID, listID)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// Delete deletes a bookmark
func (r *BookmarkRepository) Delete(ctx context.Context, userID, novelID uuid.UUID) error {
	query := `DELETE FROM bookmarks WHERE user_id = $1 AND novel_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, novelID)
	return err
}

// List retrieves bookmarks with filters
func (r *BookmarkRepository) List(ctx context.Context, filter models.BookmarksFilter, lang string) (*models.BookmarksResponse, error) {
	// Get lists first
	lists, err := r.GetOrCreateLists(ctx, filter.UserID)
	if err != nil {
		return nil, err
	}
	
	// Build query
	baseQuery := `
		FROM bookmarks b
		JOIN novels n ON b.novel_id = n.id
		JOIN novel_localizations nl ON n.id = nl.novel_id AND nl.lang = $2
		LEFT JOIN reading_progress rp ON b.user_id = rp.user_id AND b.novel_id = rp.novel_id
		LEFT JOIN chapters c ON rp.chapter_id = c.id
		LEFT JOIN (
			SELECT novel_id, MAX(number) as max_chapter, MAX(published_at) as latest_published
			FROM chapters
			WHERE published_at IS NOT NULL
			GROUP BY novel_id
		) lc ON n.id = lc.novel_id
		WHERE b.user_id = $1`
	
	args := []interface{}{filter.UserID, lang}
	argIndex := 3
	
	// Filter by list
	if filter.ListCode != nil {
		// Get list ID by code
		list, err := r.GetListByCode(ctx, filter.UserID, *filter.ListCode)
		if err != nil {
			return nil, err
		}
		if list != nil {
			baseQuery += fmt.Sprintf(" AND b.list_id = $%d", argIndex)
			args = append(args, list.ID)
			argIndex++
		}
	}
	
	// Count total
	var totalCount int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err = r.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, err
	}
	
	// Apply sorting
	var orderBy string
	switch filter.Sort {
	case "date_added":
		orderBy = "b.created_at DESC"
	case "title":
		orderBy = "nl.title ASC"
	default: // latest_update
		orderBy = "COALESCE(lc.latest_published, b.created_at) DESC"
	}
	
	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`
		SELECT 
			b.id, b.user_id, b.novel_id, b.list_id, b.created_at, b.updated_at,
			n.id as novel_id, n.slug, n.cover_image_key, n.translation_status,
			nl.title as novel_title,
			(SELECT COUNT(*) FROM chapters WHERE novel_id = n.id AND published_at IS NOT NULL) as chapters_count,
			COALESCE(n.rating, 0) as rating,
			rp.chapter_id as progress_chapter_id,
			c.number as progress_chapter_num,
			lc.max_chapter as total_chapters,
			rp.updated_at as progress_updated_at,
			CASE WHEN c.number IS NOT NULL AND lc.max_chapter > c.number THEN true ELSE false END as has_new
		%s
		ORDER BY %s
		LIMIT %d OFFSET %d`,
		baseQuery, orderBy, filter.Limit, offset)
	
	rows, err := r.db.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	bookmarks := make([]models.Bookmark, 0)
	for rows.Next() {
		var b models.Bookmark
		var novel models.BookmarkNovel
		var progressChapterID *uuid.UUID
		var progressChapterNum *int
		var totalChapters *int
		var progressUpdatedAt *string
		var hasNew bool
		
		err := rows.Scan(
			&b.ID, &b.UserID, &b.NovelID, &b.ListID, &b.CreatedAt, &b.UpdatedAt,
			&novel.ID, &novel.Slug, &novel.CoverImageKey, &novel.TranslationStatus,
			&novel.Title,
			&novel.ChaptersCount,
			&novel.Rating,
			&progressChapterID,
			&progressChapterNum,
			&totalChapters,
			&progressUpdatedAt,
			&hasNew,
		)
		if err != nil {
			return nil, err
		}
		
		b.Novel = &novel
		b.HasNewChapter = hasNew
		
		if progressChapterID != nil {
			b.Progress = &models.BookmarkReadingProgress{
				ChapterID: *progressChapterID,
			}
			if progressChapterNum != nil {
				b.Progress.ChapterNum = *progressChapterNum
			}
			if totalChapters != nil {
				b.Progress.TotalChapters = *totalChapters
			}
		}
		
		bookmarks = append(bookmarks, b)
	}
	
	return &models.BookmarksResponse{
		Bookmarks:  bookmarks,
		Lists:      lists,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

// GetStats gets bookmark statistics for a user
func (r *BookmarkRepository) GetStats(ctx context.Context, userID uuid.UUID) ([]models.BookmarkListStats, error) {
	query := `
		SELECT bl.code, COUNT(b.id) as count
		FROM bookmark_lists bl
		LEFT JOIN bookmarks b ON bl.id = b.list_id
		WHERE bl.user_id = $1
		GROUP BY bl.code
		ORDER BY bl.sort_order`
	
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stats := make([]models.BookmarkListStats, 0)
	for rows.Next() {
		var stat models.BookmarkListStats
		if err := rows.Scan(&stat.ListCode, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

// GetNovelBookmarkStatus gets bookmark status for a novel (for display on novel page)
func (r *BookmarkRepository) GetNovelBookmarkStatus(ctx context.Context, userID, novelID uuid.UUID) (*models.BookmarkListCode, error) {
	query := `
		SELECT bl.code
		FROM bookmarks b
		JOIN bookmark_lists bl ON b.list_id = bl.id
		WHERE b.user_id = $1 AND b.novel_id = $2`
	
	var code models.BookmarkListCode
	err := r.db.GetContext(ctx, &code, query, userID, novelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &code, nil
}
