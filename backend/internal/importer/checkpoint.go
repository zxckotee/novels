package importer

import "github.com/google/uuid"

// Checkpoint is persisted in import_runs.checkpoint to allow pause/resume.
// NextIndex is 0-based index of the next chapter to import.
type Checkpoint struct {
	NovelID     uuid.UUID `json:"novelId"`
	Slug        string    `json:"slug"`
	NextIndex   int       `json:"nextIndex"`
	TotalChapters int     `json:"totalChapters"`
}

