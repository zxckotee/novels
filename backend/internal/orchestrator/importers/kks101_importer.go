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

type Kks101Importer struct{}

func (Kks101Importer) Name() string { return "101kks" }

func (Kks101Importer) CanImport(originalLink string) bool {
	u, err := url.Parse(strings.TrimSpace(originalLink))
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return host == "101kks.com" || strings.HasSuffix(host, ".101kks.com")
}

func (Kks101Importer) Import(ctx context.Context, db *sqlx.DB, proposal *models.NovelProposal, uploadsDir string, checkpoint *importer.Checkpoint, onChapter func(cp *importer.Checkpoint, chaptersSaved int) error, cookieHeader string) (uuid.UUID, *importer.Checkpoint, error) {
	storageState := strings.TrimSpace(os.Getenv("KKS101_STORAGE_STATE"))
	if storageState == "" {
		storageState = "/app/cookies/101kks_storage.json"
	}
	referer := strings.TrimSpace(os.Getenv("KKS101_REFERER"))

	res, cp, err := importer.Import101KksResumable(ctx, db, importer.Import101KksOptions{
		PageURL:          proposal.OriginalLink,
		ChaptersLimit:    0,
		UploadDir:        uploadsDir,
		StorageStatePath: storageState,
		Referer:          referer,
		Cookie:           strings.TrimSpace(cookieHeader),
	}, checkpoint, onChapter)
	if err != nil {
		return uuid.Nil, cp, err
	}
	return res.NovelID, cp, nil
}

