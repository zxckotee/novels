package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"novels/internal/domain/models"
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
			id, parent_id, root_id, depth, target_type, target_id,
			user_id, body, is_deleted, is_spoiler, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, false, $9, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		comment.ID,
		comment.ParentID,
		comment.RootID,
		comment.Depth,
		comment.TargetType,
		comment.TargetID,
		comment.UserID,
		comment.Body,
		comment.IsSpoiler,
	).Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
}

// GetByID retrieves a comment by ID
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	query := `
		SELECT 
			c.id, c.parent_id, c.root_id, c.depth,
			c.target_type, c.target_id, c.user_id, c.body,
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
			c.target_type, c.target_id, c.user_id, c.body,
			c.is_deleted, c.is_spoiler,
			c.likes_count, c.dislikes_count, c.replies_count,
			c.created_at, c.updated_at,
			u.id as "user.id",
			COALESCE(up.display_name, u.email) as "user.display_name",
			up.avatar_url as "user.avatar_url",
			COALESCE(ux.level, 1) as "user.level",
			COALESCE(ur.role, 'user') as "user.role"
		FROM comments c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN user_profiles up ON u.id = up.user_id
		LEFT JOIN user_xp ux ON u.id = ux.user_id
		LEFT JOIN user_roles ur ON u.id = ur.user_id
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
		&comment.TargetType, &comment.TargetID, &comment.UserID, &comment.Body,
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
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		WHERE c.target_type = $1 AND c.target_id = $2`

	args := []interface{}{filter.TargetType, filter.TargetID}
	argIndex := 3

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
			c.target_type, c.target_id, c.user_id, c.body,
			c.is_deleted, c.is_spoiler,
			c.likes_count, c.dislikes_count, c.replies_count,
			c.created_at, c.updated_at,
			u.id as user_id,
			COALESCE(up.display_name, u.email) as user_display_name,
			up.avatar_url as user_avatar_url,
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
			&comment.TargetType, &comment.TargetID, &comment.UserID, &comment.Body,
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
		SET body = $2, is_spoiler = $3, updated_at = NOW()
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
		SET is_deleted = true, body = '[удалено]', updated_at = NOW()
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

			// Update counts
			if value == 1 {
				_, err = tx.ExecContext(ctx, `UPDATE comments SET likes_count = likes_count - 1 WHERE id = $1`, commentID)
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE comments SET dislikes_count = dislikes_count - 1 WHERE id = $1`, commentID)
			}
		} else {
			// Update vote
			_, err = tx.ExecContext(ctx, `UPDATE comment_votes SET value = $3 WHERE comment_id = $1 AND user_id = $2`, commentID, userID, value)
			if err != nil {
				return err
			}

			// Update counts (swap)
			if value == 1 {
				_, err = tx.ExecContext(ctx, `UPDATE comments SET likes_count = likes_count + 1, dislikes_count = dislikes_count - 1 WHERE id = $1`, commentID)
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE comments SET likes_count = likes_count - 1, dislikes_count = dislikes_count + 1 WHERE id = $1`, commentID)
			}
		}
	} else {
		// Insert new vote
		_, err = tx.ExecContext(ctx, `INSERT INTO comment_votes (comment_id, user_id, value, created_at) VALUES ($1, $2, $3, NOW())`, commentID, userID, value)
		if err != nil {
			return err
		}

		// Update counts
		if value == 1 {
			_, err = tx.ExecContext(ctx, `UPDATE comments SET likes_count = likes_count + 1 WHERE id = $1`, commentID)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE comments SET dislikes_count = dislikes_count + 1 WHERE id = $1`, commentID)
		}
	}

	if err != nil {
		return err
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
