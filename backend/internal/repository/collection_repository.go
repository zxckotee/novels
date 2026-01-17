package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"novels-backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CollectionRepository handles collection database operations
type CollectionRepository struct {
	db *sqlx.DB
}

// NewCollectionRepository creates a new collection repository
func NewCollectionRepository(db *sqlx.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

// Create creates a new collection
func (r *CollectionRepository) Create(ctx context.Context, collection *models.Collection) error {
	query := `
		INSERT INTO collections (user_id, slug, title, description, cover_url, is_public)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowxContext(ctx, query,
		collection.UserID,
		collection.Slug,
		collection.Title,
		collection.Description,
		collection.CoverURL,
		collection.IsPublic,
	).Scan(&collection.ID, &collection.CreatedAt, &collection.UpdatedAt)
}

// GetByID retrieves a collection by ID
func (r *CollectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Collection, error) {
	var collection models.Collection
	query := `SELECT * FROM collections WHERE id = $1`
	err := r.db.GetContext(ctx, &collection, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &collection, err
}

// GetBySlug retrieves a collection by user ID and slug
func (r *CollectionRepository) GetBySlug(ctx context.Context, userID uuid.UUID, slug string) (*models.Collection, error) {
	var collection models.Collection
	query := `SELECT * FROM collections WHERE user_id = $1 AND slug = $2`
	err := r.db.GetContext(ctx, &collection, query, userID, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &collection, err
}

// Update updates a collection
func (r *CollectionRepository) Update(ctx context.Context, collection *models.Collection) error {
	query := `
		UPDATE collections 
		SET title = $2, description = $3, cover_url = $4, is_public = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowxContext(ctx, query,
		collection.ID,
		collection.Title,
		collection.Description,
		collection.CoverURL,
		collection.IsPublic,
	).Scan(&collection.UpdatedAt)
}

// Delete deletes a collection
func (r *CollectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM collections WHERE id = $1`, id)
	return err
}

// List lists collections with filters
func (r *CollectionRepository) List(ctx context.Context, params models.CollectionListParams) ([]models.CollectionCard, int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, "is_public = true")

	if params.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, *params.UserID)
		argNum++
	}

	if params.IsFeatured != nil && *params.IsFeatured {
		conditions = append(conditions, "is_featured = true")
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Order
	orderBy := "ORDER BY "
	switch params.Sort {
	case "votes":
		orderBy += "votes_count DESC"
	case "recent":
		orderBy += "created_at DESC"
	default: // popular
		orderBy += "votes_count DESC, views_count DESC"
	}

	// Count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM collections %s`, whereClause)
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

	// Main query
	query := fmt.Sprintf(`
		SELECT id, user_id, slug, title, description, cover_url, votes_count, items_count, created_at
		FROM collections
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	var collections []models.CollectionCard
	if err := r.db.SelectContext(ctx, &collections, query, args...); err != nil {
		return nil, 0, err
	}

	return collections, total, nil
}

// GetUserCollections gets collections for a specific user
func (r *CollectionRepository) GetUserCollections(ctx context.Context, userID uuid.UUID, includePrivate bool) ([]models.CollectionCard, error) {
	var collections []models.CollectionCard
	query := `
		SELECT id, user_id, slug, title, description, cover_url, votes_count, items_count, created_at
		FROM collections
		WHERE user_id = $1`

	if !includePrivate {
		query += " AND is_public = true"
	}

	query += " ORDER BY created_at DESC"

	err := r.db.SelectContext(ctx, &collections, query, userID)
	return collections, err
}

// AddItem adds a novel to a collection
func (r *CollectionRepository) AddItem(ctx context.Context, item *models.CollectionItem) error {
	query := `
		INSERT INTO collection_items (collection_id, novel_id, position, note)
		VALUES ($1, $2, COALESCE($3, (SELECT COALESCE(MAX(position), 0) + 1 FROM collection_items WHERE collection_id = $1)), $4)
		RETURNING id, position, added_at`

	return r.db.QueryRowxContext(ctx, query,
		item.CollectionID,
		item.NovelID,
		item.Position,
		item.Note,
	).Scan(&item.ID, &item.Position, &item.AddedAt)
}

// RemoveItem removes a novel from a collection
func (r *CollectionRepository) RemoveItem(ctx context.Context, collectionID, novelID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM collection_items WHERE collection_id = $1 AND novel_id = $2`,
		collectionID, novelID)
	return err
}

// UpdateItem updates a collection item
func (r *CollectionRepository) UpdateItem(ctx context.Context, collectionID, novelID uuid.UUID, position *int, note *string) error {
	var sets []string
	var args []interface{}
	argNum := 1

	if position != nil {
		sets = append(sets, fmt.Sprintf("position = $%d", argNum))
		args = append(args, *position)
		argNum++
	}

	if note != nil {
		sets = append(sets, fmt.Sprintf("note = $%d", argNum))
		args = append(args, *note)
		argNum++
	}

	if len(sets) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE collection_items SET %s
		WHERE collection_id = $%d AND novel_id = $%d`,
		strings.Join(sets, ", "), argNum, argNum+1)

	args = append(args, collectionID, novelID)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetItems gets all items in a collection
func (r *CollectionRepository) GetItems(ctx context.Context, collectionID uuid.UUID) ([]models.CollectionItem, error) {
	var items []models.CollectionItem
	query := `
		SELECT
			ci.*,
			n.slug as novel_slug,
			COALESCE(nl.title, n.slug) as novel_title,
			CASE
				WHEN n.cover_image_key IS NOT NULL AND n.cover_image_key != '' THEN '/uploads/' || n.cover_image_key
				ELSE NULL
			END as novel_cover_url,
			COALESCE(n.rating_sum::float / NULLIF(n.rating_count, 0), 0) as novel_rating,
			nl.description as novel_description,
			n.translation_status as novel_translation_status
		FROM collection_items ci
		JOIN novels n ON ci.novel_id = n.id
		LEFT JOIN novel_localizations nl ON nl.novel_id = n.id AND nl.lang = 'ru'
		WHERE ci.collection_id = $1
		ORDER BY ci.position`

	err := r.db.SelectContext(ctx, &items, query, collectionID)
	return items, err
}

// ReorderItems reorders items in a collection
func (r *CollectionRepository) ReorderItems(ctx context.Context, collectionID uuid.UUID, items []struct {
	NovelID  uuid.UUID
	Position int
}) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err := tx.ExecContext(ctx,
			`UPDATE collection_items SET position = $1 WHERE collection_id = $2 AND novel_id = $3`,
			item.Position, collectionID, item.NovelID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Vote adds or removes a vote on a collection
func (r *CollectionRepository) Vote(ctx context.Context, vote *models.CollectionVote) error {
	query := `
		INSERT INTO collection_votes (collection_id, user_id, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (collection_id, user_id) DO UPDATE SET value = $3
		RETURNING id, created_at`

	return r.db.QueryRowxContext(ctx, query,
		vote.CollectionID,
		vote.UserID,
		vote.Value,
	).Scan(&vote.ID, &vote.CreatedAt)
}

// RemoveVote removes a vote from a collection
func (r *CollectionRepository) RemoveVote(ctx context.Context, collectionID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM collection_votes WHERE collection_id = $1 AND user_id = $2`,
		collectionID, userID)
	return err
}

// GetUserVote gets user's vote on a collection
func (r *CollectionRepository) GetUserVote(ctx context.Context, collectionID, userID uuid.UUID) (*models.CollectionVote, error) {
	var vote models.CollectionVote
	query := `SELECT * FROM collection_votes WHERE collection_id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &vote, query, collectionID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &vote, err
}

// IncrementViews increments view count
func (r *CollectionRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE collections SET views_count = views_count + 1 WHERE id = $1`, id)
	return err
}

// GetPreviewCovers gets first N novel covers for collection preview
func (r *CollectionRepository) GetPreviewCovers(ctx context.Context, collectionID uuid.UUID, limit int) ([]string, error) {
	var covers []string
	query := `
		SELECT '/uploads/' || n.cover_image_key
		FROM collection_items ci
		JOIN novels n ON ci.novel_id = n.id
		WHERE ci.collection_id = $1 AND n.cover_image_key IS NOT NULL AND n.cover_image_key != ''
		ORDER BY ci.position
		LIMIT $2`

	err := r.db.SelectContext(ctx, &covers, query, collectionID, limit)
	return covers, err
}

// SetFeatured sets or unsets featured status (admin only)
func (r *CollectionRepository) SetFeatured(ctx context.Context, id uuid.UUID, featured bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE collections SET is_featured = $2 WHERE id = $1`, id, featured)
	return err
}

// GetFeaturedCollections gets featured collections for homepage
func (r *CollectionRepository) GetFeaturedCollections(ctx context.Context, limit int) ([]models.CollectionCard, error) {
	var collections []models.CollectionCard
	query := `
		SELECT id, user_id, slug, title, description, cover_url, votes_count, items_count, created_at
		FROM collections
		WHERE is_featured = true AND is_public = true
		ORDER BY votes_count DESC
		LIMIT $1`

	err := r.db.SelectContext(ctx, &collections, query, limit)
	return collections, err
}

// GetPopularCollections gets most popular collections
func (r *CollectionRepository) GetPopularCollections(ctx context.Context, limit int) ([]models.CollectionCard, error) {
	var collections []models.CollectionCard
	query := `
		SELECT id, user_id, slug, title, description, cover_url, votes_count, items_count, created_at
		FROM collections
		WHERE is_public = true AND items_count >= 3
		ORDER BY votes_count DESC, views_count DESC
		LIMIT $1`

	err := r.db.SelectContext(ctx, &collections, query, limit)
	return collections, err
}
