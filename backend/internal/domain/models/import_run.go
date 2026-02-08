package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ImportRunStatus string

const (
	ImportRunStatusRunning       ImportRunStatus = "running"
	ImportRunStatusPauseRequested ImportRunStatus = "pause_requested"
	ImportRunStatusPaused        ImportRunStatus = "paused"
	ImportRunStatusSucceeded     ImportRunStatus = "succeeded"
	ImportRunStatusFailed        ImportRunStatus = "failed"
	ImportRunStatusCancelled     ImportRunStatus = "cancelled"
)

type ImportRun struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	ProposalID uuid.UUID      `json:"proposalId" db:"proposal_id"`
	NovelID   *uuid.UUID      `json:"novelId,omitempty" db:"novel_id"`
	Importer  string          `json:"importer" db:"importer"`
	Status    ImportRunStatus `json:"status" db:"status"`
	Error     *string         `json:"error,omitempty" db:"error"`

	ProgressCurrent int            `json:"progressCurrent" db:"progress_current"`
	ProgressTotal   int            `json:"progressTotal" db:"progress_total"`
	Checkpoint      json.RawMessage `json:"checkpoint,omitempty" db:"checkpoint"`

	CloudflareBlocked bool `json:"cloudflareBlocked" db:"cloudflare_blocked"`

	StartedAt  time.Time  `json:"startedAt" db:"started_at"`
	FinishedAt *time.Time `json:"finishedAt,omitempty" db:"finished_at"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
}

type ImportRunCookie struct {
	RunID       uuid.UUID `json:"runId" db:"run_id"`
	CookieHeader string   `json:"cookieHeader" db:"cookie_header"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}