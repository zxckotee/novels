package importers

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/importer"
)

type FanqieImporter struct{}

func (FanqieImporter) Name() string { return "fanqie" }

func (FanqieImporter) CanImport(originalLink string) bool {
	u, err := url.Parse(strings.TrimSpace(originalLink))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "fanqienovel.com" || strings.HasSuffix(host, ".fanqienovel.com")
}

func (FanqieImporter) Import(ctx context.Context, db *sqlx.DB, proposal *models.NovelProposal, uploadsDir string) (uuid.UUID, error) {
	res, err := importer.ImportFanqie(ctx, db, importer.ImportFanqieOptions{
		PageURL:       proposal.OriginalLink,
		ChaptersLimit: 0,
		UploadDir:     uploadsDir,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return res.NovelID, nil
}

