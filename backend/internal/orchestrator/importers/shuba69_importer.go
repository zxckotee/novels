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

type Shuba69Importer struct{}

func (Shuba69Importer) Name() string { return "69shuba" }

func (Shuba69Importer) CanImport(originalLink string) bool {
	u, err := url.Parse(strings.TrimSpace(originalLink))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "www.69shuba.com" || host == "69shuba.com" || strings.HasSuffix(host, ".69shuba.com")
}

func (Shuba69Importer) Import(ctx context.Context, db *sqlx.DB, proposal *models.NovelProposal, uploadsDir string, checkpoint *importer.Checkpoint, onChapter func(cp *importer.Checkpoint, chaptersSaved int) error, cookieHeader string) (uuid.UUID, *importer.Checkpoint, error) {
	// If present, reuse user-provided interactive session cookies exported via tools/shuba-browser.
	storageState := strings.TrimSpace(os.Getenv("SHUBA_STORAGE_STATE"))
	if storageState == "" {
		storageState = "/app/cookies/69shuba_storage.json"
	}

	res, cp, err := importer.Import69ShubaResumable(ctx, db, importer.Import69ShubaOptions{
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

