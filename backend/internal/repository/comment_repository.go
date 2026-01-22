package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels-backend/internal/domain/models"
)

type CommentRepository struct {
	db *sqlx.DB
}

func NewCommentRepository(db *sqlx.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (
			id, parent_id, root_id, depth, target_type, target_id, anchor,
			user_id, body, content, is_deleted, is_spoiler, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, false, $11, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		comment.ID,
		comment.ParentID,
		comment.RootID,
		comment.Depth,
		comment.TargetType,
		comment.TargetID,
		comment.Anchor,
		comment.UserID,
		comment.Body,
		comment.Body, // legacy column (NOT NULL in older schema)
		comment.IsSpoiler,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	query := `
		SELECT 
			c.id, c.parent_id, c.root_id, c.depth,
			c.target_type, c.target_id, c.anchor, c.user_id, c.body,
			c.is_deleted, c.is_spoiler,
			c.likes_count, c.dislikes_count, c.replies_count,
			c.created_at, c.updated_at
		FROM comments c
		WHERE c.id = $1`

	err := r.db.GetContext(ctx, &comment, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &comment, nil
}

// GetByIDWithUser retrieves a comment with user info
func (r *CommentRepository) GetByIDWithUser(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*models.Comment, error) {
	query := `
		SELECT 
			c.id, c.parent_id, c.root_id, c.depth,
			c.target_type, c.target_id, c.anchor, c.user_id, c.body,
			c.is_deleted, c.is_spoiler,
			c.likes_count, c.dislikes_count, c.replies_count,
			c.created_at, c.updated_at,
			u.id as "user.id",
			COALESCE(up.display_name, u.email) as "user.display_name",
			CASE WHEN up.avatar_key IS NOT NULL THEN '/uploads/' || up.avatar_key ELSE NULL END as "user.avatar_url",
			COALESCE(ux.level, 1) as "user.level",
			COALESCE(ur.role, 'user') as "user.role"
		FROM comments c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		LEFT JOIN user_xp ux ON u.id = ux.user_id
		LEFT JOIN LATERAL (
			SELECT role
			FROM user_roles
			WHERE user_id = u.id
			ORDER BY CASE role
				WHEN 'admin' THEN 4
				WHEN 'moderator' THEN 3
				WHEN 'premium' THEN 2
				ELSE 1
			END DESC
			LIMIT 1
		) ur ON true
		WHERE c.id = $1`

	rows, err := r.db.QueryxContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var comment models.Comment
	var user models.CommentUser

	err = rows.Scan(
		&comment.ID, &comment.ParentID, &comment.RootID, &comment.Depth,
		&comment.TargetType, &comment.TargetID, &comment.Anchor, &comment.UserID, &comment.Body,
		&comment.IsDeleted, &comment.IsSpoiler,
		&comment.LikesCount, &comment.DislikesCount, &comment.RepliesCount,
		&comment.CreatedAt, &comment.UpdatedAt,
		&user.ID, &user.DisplayName, &user.AvatarURL, &user.Level, &user.Role,
	)
	if err != nil {
		return nil, err
	}

	comment.User = &user

	// Get viewer's vote if logged in
	if viewerID != nil {
		vote, err := r.GetUserVote(ctx, id, *viewerID)
		if err == nil && vote != nil {
			comment.UserVote = &vote.Value
		}
	}

	return &comment, nil
}

// List retrieves comments with filters
func (r *CommentRepository) List(ctx context.Context, filter models.CommentsFilter, viewerID *uuid.UUID) (*models.CommentsResponse, error) {
	// Build query
	baseQuery := `
		FROM comments c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		LEFT JOIN user_xp ux ON u.id = ux.user_id
		LEFT JOIN LATERAL (
			SELECT role
			FROM user_roles
			WHERE user_id = u.id
			ORDER BY CASE role
				WHEN 'admin' THEN 4
				WHEN 'moderator' THEN 3
				WHEN 'premium' THEN 2
				ELSE 1
			END DESC
			LIMIT 1
		) ur ON true
		WHERE c.target_type = $1 AND c.target_id = $2`

	args := []interface{}{filter.TargetType, filter.TargetID}
	argIndex := 3

	// Filter by anchor (nil = any, set = exact match, including replies if parent_id is set)
	if filter.Anchor != nil {
		baseQuery += fmt.Sprintf(" AND c.anchor = $%d", argIndex)
		args = append(args, *filter.Anchor)
		argIndex++
	}

	// Filter by parent (nil = root comments)
	if filter.ParentID == nil {
		baseQuery += " AND c.parent_id IS NULL"
	} else {
		baseQuery += fmt.Sprintf(" AND c.parent_id = $%d", argIndex)
		args = append(args, *filter.ParentID)
		argIndex++
	}

	// Filter by user
	if filter.UserID != nil {
		baseQuery += fmt.Sprintf(" AND c.user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	// Count total
	var totalCount int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Apply sorting
	var orderBy string
	switch filter.Sort {
	case "oldest":
		orderBy = "c.created_at ASC"
	case "top":
		orderBy = "(c.likes_count - c.dislikes_count) DESC, c.created_at DESC"
	default: // newest
		orderBy = "c.created_at DESC"
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.Limit
	selectQuery := fmt.Sprintf(`
		SELECT 
			c.id, c.parent_id, c.root_id, c.depth,
			c.target_type, c.target_id, c.anchor, c.user_id, c.body,
			c.is_deleted, c.is_spoiler,
			c.likes_count, c.dislikes_count, c.replies_count,
			c.created_at, c.updated_at,
			u.id as user_id,
			COALESCE(up.display_name, u.email) as user_display_name,
			CASE WHEN up.avatar_key IS NOT NULL THEN '/uploads/' || up.avatar_key ELSE NULL END as user_avatar_url,
			COALESCE(ux.level, 1) as user_level,
			COALESCE(ur.role, 'user') as user_role
		%s
		ORDER BY %s
		LIMIT %d OFFSET %d`,
		baseQuery, orderBy, filter.Limit, offset)

	rows, err := r.db.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]models.Comment, 0)
	for rows.Next() {
		var comment models.Comment
		var userID uuid.UUID
		var displayName string
		var avatarURL *string
		var level int
		var role models.UserRole

		err := rows.Scan(
			&comment.ID, &comment.ParentID, &comment.RootID, &comment.Depth,
			&comment.TargetType, &comment.TargetID, &comment.Anchor, &comment.UserID, &comment.Body,
			&comment.IsDeleted, &comment.IsSpoiler,
			&comment.LikesCount, &comment.DislikesCount, &comment.RepliesCount,
			&comment.CreatedAt, &comment.UpdatedAt,
			&userID, &displayName, &avatarURL, &level, &role,
		)
		if err != nil {
			return nil, err
		}

		comment.User = &models.CommentUser{
			ID:          userID,
			DisplayName: displayName,
			AvatarURL:   avatarURL,
			Level:       level,
			Role:        role,
		}

		comments = append(comments, comment)
	}

	// Get viewer's votes if logged in
	if viewerID != nil && len(comments) > 0 {
		commentIDs := make([]uuid.UUID, len(comments))
		for i, c := range comments {
			commentIDs[i] = c.ID
		}

		votes, err := r.GetUserVotes(ctx, commentIDs, *viewerID)
		if err == nil {
			voteMap := make(map[uuid.UUID]int)
			for _, v := range votes {
				voteMap[v.CommentID] = v.Value
			}
			for i := range comments {
				if vote, ok := voteMap[comments[i].ID]; ok {
					comments[i].UserVote = &vote
				}
			}
		}
	}

	return &models.CommentsResponse{
		Comments:   comments,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
	}, nil
}

// Update updates a comment
func (r *CommentRepository) Update(ctx context.Context, id uuid.UUID, body string, isSpoiler bool) error {
	query := `
		UPDATE comments 
		SET body = $2, content = $2, is_spoiler = $3, updated_at = NOW()
		WHERE id = $1 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query, id, body, isSpoiler)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete soft-deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE comments 
		SET is_deleted = true, body = '[удалено]', content = '[удалено]', updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// HardDelete permanently deletes a comment (admin only)
func (r *CommentRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Vote adds or updates a vote on a comment
func (r *CommentRepository) Vote(ctx context.Context, commentID, userID uuid.UUID, value int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get existing vote
	var existingVote *models.CommentVote
	query := `SELECT comment_id, user_id, value, created_at FROM comment_votes WHERE comment_id = $1 AND user_id = $2`
	var vote models.CommentVote
	err = tx.GetContext(ctx, &vote, query, commentID, userID)
	if err == nil {
		existingVote = &vote
	} else if err != sql.ErrNoRows {
		return err
	}

	if existingVote != nil {
		if existingVote.Value == value {
			// Remove vote (toggle off)
			_, err = tx.ExecContext(ctx, `DELETE FROM comment_votes WHERE comment_id = $1 AND user_id = $2`, commentID, userID)
			if err != nil {
				return err
			}
		} else {
			// Update vote
			_, err = tx.ExecContext(ctx, `UPDATE comment_votes SET value = $3 WHERE comment_id = $1 AND user_id = $2`, commentID, userID, value)
			if err != nil {
				return err
			}
		}
	} else {
		// Insert new vote
		_, err = tx.ExecContext(ctx, `INSERT INTO comment_votes (comment_id, user_id, value, created_at) VALUES ($1, $2, $3, NOW())`, commentID, userID, value)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetUserVote gets a user's vote on a comment
func (r *CommentRepository) GetUserVote(ctx context.Context, commentID, userID uuid.UUID) (*models.CommentVote, error) {
	var vote models.CommentVote
	query := `SELECT comment_id, user_id, value, created_at FROM comment_votes WHERE comment_id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &vote, query, commentID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// GetUserVotes gets a user's votes on multiple comments
func (r *CommentRepository) GetUserVotes(ctx context.Context, commentIDs []uuid.UUID, userID uuid.UUID) ([]models.CommentVote, error) {
	if len(commentIDs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`SELECT comment_id, user_id, value, created_at FROM comment_votes WHERE comment_id IN (?) AND user_id = ?`, commentIDs, userID)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var votes []models.CommentVote
	err = r.db.SelectContext(ctx, &votes, query, args...)
	return votes, err
}

// Report creates a report for a comment
func (r *CommentRepository) Report(ctx context.Context, report *models.CommentReport) error {
	query := `
		INSERT INTO comment_reports (id, comment_id, user_id, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'pending', NOW(), NOW())
		ON CONFLICT (comment_id, user_id) DO UPDATE SET reason = $4, updated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query, report.ID, report.CommentID, report.UserID, report.Reason)
	return err
}

// UpdateRepliesCount updates the replies count for a parent comment
func (r *CommentRepository) UpdateRepliesCount(ctx context.Context, parentID uuid.UUID) error {
	query := `
		UPDATE comments 
		SET replies_count = (SELECT COUNT(*) FROM comments WHERE parent_id = $1)
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, parentID)
	return err
}

// GetReplies gets replies to a comment
func (r *CommentRepository) GetReplies(ctx context.Context, parentID uuid.UUID, limit int, viewerID *uuid.UUID) ([]models.Comment, error) {
	filter := models.CommentsFilter{
		ParentID: &parentID,
		Sort:     "oldest",
		Page:     1,
		Limit:    limit,
	}

	// Get parent comment to know target
	parent, err := r.GetByID(ctx, parentID)
	if err != nil || parent == nil {
		return nil, err
	}

	filter.TargetType = parent.TargetType
	filter.TargetID = parent.TargetID

	response, err := r.List(ctx, filter, viewerID)
	if err != nil {
		return nil, err
	}

	return response.Comments, nil
}

// ========================
// ADMIN METHODS
// ========================

// AdminListComments получает список комментариев для админа с расширенными фильтрами
func (r *CommentRepository) AdminListComments(ctx context.Context, filter models.AdminCommentsFilter) ([]models.Comment, int, error) {
	baseQuery := `FROM comments c`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.TargetType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("c.target_type = $%d", argIndex))
		args = append(args, filter.TargetType)
		argIndex++
	}

	if filter.TargetID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("c.target_id = $%d", argIndex))
		args = append(args, *filter.TargetID)
		argIndex++
	}

	if filter.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("c.user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.IsDeleted != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("c.is_deleted = $%d", argIndex))
		args = append(args, *filter.IsDeleted)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) "+baseQuery+" "+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "c.created_at DESC"
	if filter.Sort == "oldest" {
		orderBy = "c.created_at ASC"
	} else if filter.Sort == "reports" {
		orderBy = "(SELECT COUNT(*) FROM comment_reports WHERE comment_id = c.id) DESC"
	}

	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.parent_id, c.root_id, c.depth,
		       c.target_type, c.target_id, c.user_id, c.body,
		       c.is_deleted, c.is_spoiler,
		       c.likes_count, c.dislikes_count, c.replies_count,
		       c.created_at, c.updated_at
		%s %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, argIndex, argIndex+1)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var comments []models.Comment
	err = r.db.SelectContext(ctx, &comments, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// SoftDeleteComment помечает комментарий как удаленный
func (r *CommentRepository) SoftDeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return r.Delete(ctx, commentID)
}

// HardDeleteComment полностью удаляет комментарий
func (r *CommentRepository) HardDeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return r.HardDelete(ctx, commentID)
}

// GetReports получает список жалоб с фильтрацией
func (r *CommentRepository) GetReports(ctx context.Context, filter models.ReportsFilter) ([]models.CommentReport, int, error) {
	baseQuery := `FROM comment_reports cr`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("cr.status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.CommentID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cr.comment_id = $%d", argIndex))
		args = append(args, *filter.CommentID)
		argIndex++
	}

	if filter.ReporterID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("cr.user_id = $%d", argIndex))
		args = append(args, *filter.ReporterID)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) "+baseQuery+" "+whereClause, args...)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "cr.created_at DESC"
	if filter.Sort == "oldest" {
		orderBy = "cr.created_at ASC"
	}

	selectQuery := fmt.Sprintf(`
		SELECT cr.id, cr.comment_id, cr.user_id, cr.reason, cr.status, cr.created_at, cr.updated_at
		%s %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, argIndex, argIndex+1)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var reports []models.CommentReport
	err = r.db.SelectContext(ctx, &reports, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// ResolveReport обрабатывает жалобу
func (r *CommentRepository) ResolveReport(ctx context.Context, reportID uuid.UUID, action, reason string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Получаем жалобу
	var report models.CommentReport
	err = tx.GetContext(ctx, &report, `SELECT id, comment_id, user_id, reason, status FROM comment_reports WHERE id = $1`, reportID)
	if err != nil {
		return err
	}

	// Выполняем действие
	switch action {
	case "resolve":
		_, err = tx.ExecContext(ctx, `UPDATE comment_reports SET status = 'resolved', updated_at = NOW() WHERE id = $1`, reportID)
	case "dismiss":
		_, err = tx.ExecContext(ctx, `UPDATE comment_reports SET status = 'dismissed', updated_at = NOW() WHERE id = $1`, reportID)
	case "delete_comment":
		// Удаляем комментарий
		_, err = tx.ExecContext(ctx, `UPDATE comments SET is_deleted = true, body = '[удалено]' WHERE id = $1`, report.CommentID)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE comment_reports SET status = 'resolved', updated_at = NOW() WHERE id = $1`, reportID)
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
