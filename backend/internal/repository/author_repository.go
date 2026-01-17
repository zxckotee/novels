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

// AuthorRepository репозиторий для работы с авторами
type AuthorRepository struct {
	db *sqlx.DB
}

// NewAuthorRepository создает новый AuthorRepository
func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}

// List получает список авторов с фильтрацией и пагинацией
func (r *AuthorRepository) List(ctx context.Context, filter models.AuthorsFilter) ([]models.Author, int, error) {
	baseQuery := `FROM authors a`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Фильтр по языку и поиску
	if filter.Lang != "" && filter.Query != "" {
		baseQuery += ` JOIN author_localizations al ON a.id = al.author_id AND al.lang = $1`
		whereConditions = append(whereConditions, fmt.Sprintf("al.name ILIKE $%d", argIndex+1))
		args = append(args, filter.Lang, "%"+filter.Query+"%")
		argIndex += 2
	} else if filter.Lang != "" {
		baseQuery += ` JOIN author_localizations al ON a.id = al.author_id AND al.lang = $1`
		args = append(args, filter.Lang)
		argIndex++
	} else if filter.Query != "" {
		baseQuery += ` JOIN author_localizations al ON a.id = al.author_id`
		whereConditions = append(whereConditions, fmt.Sprintf("al.name ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Query+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(DISTINCT a.id) " + baseQuery + " " + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count authors: %w", err)
	}

	// Сортировка
	orderBy := "a.created_at"
	switch filter.Sort {
	case "name":
		if filter.Lang != "" {
			orderBy = "al.name"
		}
	case "novels_count":
		orderBy = "(SELECT COUNT(*) FROM novel_authors na WHERE na.author_id = a.id)"
	}

	order := "DESC"
	if filter.Order == "asc" {
		order = "ASC"
	}

	// Main query
	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT a.id, a.slug, a.created_at, a.updated_at
		%s %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, orderBy, order, argIndex, argIndex+1)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	var authorRows []models.Author
	err = r.db.SelectContext(ctx, &authorRows, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list authors: %w", err)
	}

	// Загружаем локализации для каждого автора
	for i := range authorRows {
		locs, err := r.GetLocalizations(ctx, authorRows[i].ID)
		if err != nil {
			return nil, 0, err
		}
		authorRows[i].Localizations = locs

		// Если указан язык, добавляем name и bio для текущего языка
		if filter.Lang != "" && locs[filter.Lang] != nil {
			authorRows[i].Name = locs[filter.Lang].Name
			authorRows[i].Bio = locs[filter.Lang].Bio
		}
	}

	return authorRows, total, nil
}

// GetByID получает автора по ID
func (r *AuthorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Author, error) {
	query := `SELECT id, slug, created_at, updated_at FROM authors WHERE id = $1`

	var author models.Author
	err := r.db.GetContext(ctx, &author, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	// Загружаем локализации
	locs, err := r.GetLocalizations(ctx, author.ID)
	if err != nil {
		return nil, err
	}
	author.Localizations = locs

	return &author, nil
}

// GetBySlug получает автора по slug
func (r *AuthorRepository) GetBySlug(ctx context.Context, slug string) (*models.Author, error) {
	query := `SELECT id, slug, created_at, updated_at FROM authors WHERE slug = $1`

	var author models.Author
	err := r.db.GetContext(ctx, &author, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	// Загружаем локализации
	locs, err := r.GetLocalizations(ctx, author.ID)
	if err != nil {
		return nil, err
	}
	author.Localizations = locs

	return &author, nil
}

// GetLocalizations получает все локализации автора
func (r *AuthorRepository) GetLocalizations(ctx context.Context, authorID uuid.UUID) (map[string]*models.AuthorLocalization, error) {
	query := `SELECT author_id, lang, name, bio, created_at, updated_at 
	          FROM author_localizations 
	          WHERE author_id = $1`

	var locs []models.AuthorLocalization
	err := r.db.SelectContext(ctx, &locs, query, authorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get localizations: %w", err)
	}

	result := make(map[string]*models.AuthorLocalization)
	for i := range locs {
		result[locs[i].Lang] = &locs[i]
	}

	return result, nil
}

// Create создает нового автора
func (r *AuthorRepository) Create(ctx context.Context, req *models.CreateAuthorRequest) (*models.Author, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	author := &models.Author{
		ID:   uuid.New(),
		Slug: req.Slug,
	}

	// Вставляем автора
	query := `INSERT INTO authors (id, slug) VALUES ($1, $2) RETURNING created_at, updated_at`
	err = tx.QueryRowxContext(ctx, query, author.ID, author.Slug).Scan(&author.CreatedAt, &author.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}

	// Вставляем локализации
	author.Localizations = make(map[string]*models.AuthorLocalization)
	for _, loc := range req.Localizations {
		locQuery := `
			INSERT INTO author_localizations (author_id, lang, name, bio)
			VALUES ($1, $2, $3, $4)
			RETURNING created_at, updated_at
		`
		locRec := &models.AuthorLocalization{
			AuthorID: author.ID,
			Lang:     loc.Lang,
			Name:     loc.Name,
			Bio:      loc.Bio,
		}
		err = tx.QueryRowxContext(ctx, locQuery, author.ID, loc.Lang, loc.Name, loc.Bio).
			Scan(&locRec.CreatedAt, &locRec.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create localization: %w", err)
		}
		author.Localizations[loc.Lang] = locRec
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return author, nil
}

// Update обновляет автора
func (r *AuthorRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateAuthorRequest) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Обновляем slug если указан
	if req.Slug != "" {
		query := `UPDATE authors SET slug = $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, query, req.Slug, id)
		if err != nil {
			return fmt.Errorf("failed to update author: %w", err)
		}
	}

	// Обновляем локализации
	if len(req.Localizations) > 0 {
		for _, loc := range req.Localizations {
			locQuery := `
				INSERT INTO author_localizations (author_id, lang, name, bio)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (author_id, lang) DO UPDATE SET
					name = EXCLUDED.name,
					bio = EXCLUDED.bio
			`
			_, err = tx.ExecContext(ctx, locQuery, id, loc.Lang, loc.Name, loc.Bio)
			if err != nil {
				return fmt.Errorf("failed to update localization: %w", err)
			}
		}
	}

	return tx.Commit()
}

// Delete удаляет автора
func (r *AuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Проверяем, есть ли связанные новеллы
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM novel_authors WHERE author_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to check author novels: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete author with %d associated novels", count)
	}

	query := `DELETE FROM authors WHERE id = $1`
	_, err = r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete author: %w", err)
	}
	return nil
}

// GetNovelAuthors получает авторов новеллы
func (r *AuthorRepository) GetNovelAuthors(ctx context.Context, novelID uuid.UUID, lang string) ([]models.NovelAuthor, error) {
	query := `
		SELECT na.novel_id, na.author_id, na.is_primary, na.sort_order, na.created_at,
		       a.id, a.slug, a.created_at, a.updated_at,
		       COALESCE(al.name, a.slug) as name
		FROM novel_authors na
		JOIN authors a ON na.author_id = a.id
		LEFT JOIN author_localizations al ON a.id = al.author_id AND al.lang = $1
		WHERE na.novel_id = $2
		ORDER BY na.is_primary DESC, na.sort_order, al.name
	`

	rows, err := r.db.QueryxContext(ctx, query, lang, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get novel authors: %w", err)
	}
	defer rows.Close()

	var novelAuthors []models.NovelAuthor
	for rows.Next() {
		var na models.NovelAuthor
		var author models.Author
		err := rows.Scan(
			&na.NovelID, &na.AuthorID, &na.IsPrimary, &na.SortOrder, &na.CreatedAt,
			&author.ID, &author.Slug, &author.CreatedAt, &author.UpdatedAt,
			&author.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan novel author: %w", err)
		}
		na.Author = &author
		novelAuthors = append(novelAuthors, na)
	}

	return novelAuthors, nil
}

// UpdateNovelAuthors обновляет авторов новеллы
func (r *AuthorRepository) UpdateNovelAuthors(ctx context.Context, novelID uuid.UUID, authors []models.NovelAuthorInput) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем старые связи
	_, err = tx.ExecContext(ctx, `DELETE FROM novel_authors WHERE novel_id = $1`, novelID)
	if err != nil {
		return fmt.Errorf("failed to delete old authors: %w", err)
	}

	// Добавляем новые связи
	for _, author := range authors {
		authorID, err := uuid.Parse(author.AuthorID)
		if err != nil {
			return fmt.Errorf("invalid author ID: %w", err)
		}

		query := `
			INSERT INTO novel_authors (novel_id, author_id, is_primary, sort_order)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.ExecContext(ctx, query, novelID, authorID, author.IsPrimary, author.SortOrder)
		if err != nil {
			return fmt.Errorf("failed to add author: %w", err)
		}
	}

	return tx.Commit()
}
