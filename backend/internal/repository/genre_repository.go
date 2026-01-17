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

type GenreRepository struct {
	db *sqlx.DB
}

func NewGenreRepository(db *sqlx.DB) *GenreRepository {
	return &GenreRepository{db: db}
}

func (r *GenreRepository) List(ctx context.Context, filter models.GenresFilter) ([]models.GenreWithLocalizations, int, error) {
	baseQuery := `FROM genres g`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Lang != "" && filter.Query != "" {
		baseQuery += ` JOIN genre_localizations gl ON g.id = gl.genre_id AND gl.lang = $1`
		whereConditions = append(whereConditions, fmt.Sprintf("gl.name ILIKE $%d", argIndex+1))
		args = append(args, filter.Lang, "%"+filter.Query+"%")
		argIndex += 2
	} else if filter.Lang != "" {
		baseQuery += ` JOIN genre_localizations gl ON g.id = gl.genre_id AND gl.lang = $1`
		args = append(args, filter.Lang)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(DISTINCT g.id) "+baseQuery+" "+whereClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count genres: %w", err)
	}

	orderBy := "g.created_at"
	if filter.Sort == "name" && filter.Lang != "" {
		orderBy = "gl.name"
	}
	order := "DESC"
	if filter.Order == "asc" {
		order = "ASC"
	}

	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT g.id, g.slug, g.created_at
		%s %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, order, argIndex, argIndex+1)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var genres []models.GenreWithLocalizations
	err = r.db.SelectContext(ctx, &genres, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list genres: %w", err)
	}

	for i := range genres {
		locs, _ := r.GetLocalizations(ctx, genres[i].ID)
		genres[i].Localizations = locs
		if filter.Lang != "" && locs[filter.Lang] != nil {
			genres[i].Name = locs[filter.Lang].Name
		}
	}

	return genres, total, nil
}

func (r *GenreRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.GenreWithLocalizations, error) {
	var genre models.GenreWithLocalizations
	err := r.db.GetContext(ctx, &genre, `SELECT id, slug, created_at FROM genres WHERE id = $1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}

	locs, _ := r.GetLocalizations(ctx, genre.ID)
	genre.Localizations = locs
	return &genre, nil
}

func (r *GenreRepository) GetLocalizations(ctx context.Context, genreID uuid.UUID) (map[string]*models.GenreLocalization, error) {
	var locs []models.GenreLocalization
	err := r.db.SelectContext(ctx, &locs, `SELECT genre_id, lang, name FROM genre_localizations WHERE genre_id = $1`, genreID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.GenreLocalization)
	for i := range locs {
		result[locs[i].Lang] = &locs[i]
	}
	return result, nil
}

func (r *GenreRepository) Create(ctx context.Context, req *models.CreateGenreRequest) (*models.GenreWithLocalizations, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	genre := &models.GenreWithLocalizations{
		ID:   uuid.New(),
		Slug: req.Slug,
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO genres (id, slug) VALUES ($1, $2)`, genre.ID, genre.Slug)
	if err != nil {
		return nil, err
	}

	genre.Localizations = make(map[string]*models.GenreLocalization)
	for _, loc := range req.Localizations {
		_, err = tx.ExecContext(ctx, `INSERT INTO genre_localizations (genre_id, lang, name) VALUES ($1, $2, $3)`, genre.ID, loc.Lang, loc.Name)
		if err != nil {
			return nil, err
		}
		genre.Localizations[loc.Lang] = &models.GenreLocalization{
			GenreID: genre.ID,
			Lang:    loc.Lang,
			Name:    loc.Name,
		}
	}

	return genre, tx.Commit()
}

func (r *GenreRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateGenreRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if req.Slug != "" {
		_, err = tx.ExecContext(ctx, `UPDATE genres SET slug = $1 WHERE id = $2`, req.Slug, id)
		if err != nil {
			return err
		}
	}

	for _, loc := range req.Localizations {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO genre_localizations (genre_id, lang, name)
			VALUES ($1, $2, $3)
			ON CONFLICT (genre_id, lang) DO UPDATE SET name = EXCLUDED.name
		`, id, loc.Lang, loc.Name)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *GenreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM novel_genres WHERE genre_id = $1`, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete genre with %d associated novels", count)
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM genres WHERE id = $1`, id)
	return err
}
