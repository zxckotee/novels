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

type TagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) List(ctx context.Context, filter models.TagsFilter) ([]models.TagWithLocalizations, int, error) {
	baseQuery := `FROM tags t`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Lang != "" && filter.Query != "" {
		baseQuery += ` JOIN tag_localizations tl ON t.id = tl.tag_id AND tl.lang = $1`
		whereConditions = append(whereConditions, fmt.Sprintf("tl.name ILIKE $%d", argIndex+1))
		args = append(args, filter.Lang, "%"+filter.Query+"%")
		argIndex += 2
	} else if filter.Lang != "" {
		baseQuery += ` JOIN tag_localizations tl ON t.id = tl.tag_id AND tl.lang = $1`
		args = append(args, filter.Lang)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(DISTINCT t.id) "+baseQuery+" "+whereClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
	}

	orderBy := "t.created_at"
	if filter.Sort == "name" && filter.Lang != "" {
		orderBy = "tl.name"
	}
	order := "DESC"
	if filter.Order == "asc" {
		order = "ASC"
	}

	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT t.id, t.slug, t.created_at
		%s %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, order, argIndex, argIndex+1)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var tags []models.TagWithLocalizations
	err = r.db.SelectContext(ctx, &tags, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}

	for i := range tags {
		locs, _ := r.GetLocalizations(ctx, tags[i].ID)
		tags[i].Localizations = locs
		if filter.Lang != "" && locs[filter.Lang] != nil {
			tags[i].Name = locs[filter.Lang].Name
		}
	}

	return tags, total, nil
}

func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TagWithLocalizations, error) {
	var tag models.TagWithLocalizations
	err := r.db.GetContext(ctx, &tag, `SELECT id, slug, created_at FROM tags WHERE id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	locs, _ := r.GetLocalizations(ctx, tag.ID)
	tag.Localizations = locs
	return &tag, nil
}

func (r *TagRepository) GetLocalizations(ctx context.Context, tagID uuid.UUID) (map[string]*models.TagLocalization, error) {
	var locs []models.TagLocalization
	err := r.db.SelectContext(ctx, &locs, `SELECT tag_id, lang, name FROM tag_localizations WHERE tag_id = $1`, tagID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.TagLocalization)
	for i := range locs {
		result[locs[i].Lang] = &locs[i]
	}
	return result, nil
}

func (r *TagRepository) Create(ctx context.Context, req *models.CreateTagRequest) (*models.TagWithLocalizations, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	tag := &models.TagWithLocalizations{
		ID:   uuid.New(),
		Slug: req.Slug,
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO tags (id, slug) VALUES ($1, $2)`, tag.ID, tag.Slug)
	if err != nil {
		return nil, err
	}

	tag.Localizations = make(map[string]*models.TagLocalization)
	for _, loc := range req.Localizations {
		_, err = tx.ExecContext(ctx, `INSERT INTO tag_localizations (tag_id, lang, name) VALUES ($1, $2, $3)`, tag.ID, loc.Lang, loc.Name)
		if err != nil {
			return nil, err
		}
		tag.Localizations[loc.Lang] = &models.TagLocalization{
			TagID: tag.ID,
			Lang:  loc.Lang,
			Name:  loc.Name,
		}
	}

	return tag, tx.Commit()
}

func (r *TagRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateTagRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if req.Slug != "" {
		_, err = tx.ExecContext(ctx, `UPDATE tags SET slug = $1 WHERE id = $2`, req.Slug, id)
		if err != nil {
			return err
		}
	}

	for _, loc := range req.Localizations {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO tag_localizations (tag_id, lang, name)
			VALUES ($1, $2, $3)
			ON CONFLICT (tag_id, lang) DO UPDATE SET name = EXCLUDED.name
		`, id, loc.Lang, loc.Name)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM novel_tags WHERE tag_id = $1`, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete tag with %d associated novels", count)
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = $1`, id)
	return err
}
