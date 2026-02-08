package importers

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/importer"
)

type TaduImporter struct{}

func (TaduImporter) Name() string { return "tadu" }

func (TaduImporter) CanImport(originalLink string) bool {
	u, err := url.Parse(strings.TrimSpace(originalLink))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	// Support www.tadu.com and m.tadu.com URLs
	return host == "tadu.com" || host == "www.tadu.com" || host == "m.tadu.com" || strings.HasSuffix(host, ".tadu.com")
}

func (TaduImporter) Import(ctx context.Context, db *sqlx.DB, proposal *models.NovelProposal, uploadsDir string, checkpoint *importer.Checkpoint, onChapter func(cp *importer.Checkpoint, chaptersSaved int) error, cookieHeader string) (uuid.UUID, *importer.Checkpoint, error) {
	storageState := strings.TrimSpace(os.Getenv("TADU_STORAGE_STATE"))
	if storageState == "" {
		storageState = "/app/cookies/tadu_storage.json"
	}

	res, cp, err := importer.ImportTaduResumable(ctx, db, importer.ImportTaduOptions{
		PageURL:          proposal.OriginalLink,
		ChaptersLimit:    0,
		UploadDir:        uploadsDir,
		StorageStatePath: storageState,
		Cookie:           strings.TrimSpace(cookieHeader),
	}, checkpoint, onChapter)
	if err != nil {
		return uuid.Nil, cp, err
	}
	return res.NovelID, cp, nil
}

