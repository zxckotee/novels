package orchestrator

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/events"
	"novels-backend/internal/importer"
	"novels-backend/internal/repository"
)

// isCloudflareError checks if an error message indicates Cloudflare blocking
func isCloudflareError(errMsg string) bool {
	if errMsg == "" {
		return false
	}
	lower := strings.ToLower(errMsg)
	indicators := []string{
		"cloudflare",
		"cf_clearance",
		"403 blocked",
		"blocked by cloudflare",
		"cloudflare challenge",
		"turnstile",
	}
	for _, indicator := range indicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}

type ProposalImporter interface {
	Name() string
	CanImport(originalLink string) bool
	Import(ctx context.Context, db *sqlx.DB, proposal *models.NovelProposal, uploadsDir string, checkpoint *importer.Checkpoint, onChapter func(cp *importer.Checkpoint, chaptersSaved int) error, cookieHeader string) (uuid.UUID, *importer.Checkpoint, error)
}

type ImportOrchestrator struct {
	db         *sqlx.DB
	votingRepo *repository.VotingRepository
	importRuns *repository.ImportRunsRepository
	cookiesRepo *repository.ImportRunCookiesRepository
	bus        *events.Bus
	uploadsDir string
	importers  []ProposalImporter
	logger     zerolog.Logger

	mu            sync.Mutex
	activeCancels map[uuid.UUID]context.CancelFunc // runID -> cancel
}

func NewImportOrchestrator(
	db *sqlx.DB,
	votingRepo *repository.VotingRepository,
	importRuns *repository.ImportRunsRepository,
	cookiesRepo *repository.ImportRunCookiesRepository,
	bus *events.Bus,
	uploadsDir string,
	importers []ProposalImporter,
	logger zerolog.Logger,
) *ImportOrchestrator {
	return &ImportOrchestrator{
		db:         db,
		votingRepo: votingRepo,
		importRuns: importRuns,
		cookiesRepo: cookiesRepo,
		bus:        bus,
		uploadsDir: uploadsDir,
		importers:  importers,
		logger:     logger.With().Str("component", "import_orchestrator").Logger(),
		activeCancels: map[uuid.UUID]context.CancelFunc{},
	}
}

func (o *ImportOrchestrator) Register() {
	if o.bus == nil {
		return
	}

	o.bus.Subscribe(events.EventDailyVoteWinnerSelected, func(_ context.Context, evt events.Event) error {
		e := evt.(events.DailyVoteWinnerSelected)

		// Run import async: winner job should stay fast and deterministic.
		o.StartImportAsync(e.ProposalID)
		return nil
	})
}

// StartImportAsync starts importing a proposal in background and returns the run ID.
func (o *ImportOrchestrator) StartImportAsync(proposalID uuid.UUID) uuid.UUID {
	runID := uuid.New()
	go o.handleImportRun(runID, proposalID, "")
	return runID
}

// StartImportAsyncWithCookies starts importing a proposal with custom cookies in background and returns the run ID.
func (o *ImportOrchestrator) StartImportAsyncWithCookies(proposalID uuid.UUID, cookieHeader string) uuid.UUID {
	runID := uuid.New()
	go o.handleImportRun(runID, proposalID, cookieHeader)
	return runID
}

// ResumeImportAsync resumes a paused import run by its run ID.
func (o *ImportOrchestrator) ResumeImportAsync(runID uuid.UUID) {
	if o.importRuns == nil {
		return
	}
	run, err := o.importRuns.GetByID(context.Background(), runID)
	if err != nil || run == nil {
		return
	}
	// Only resume paused/pause_requested runs.
	if run.Status != models.ImportRunStatusPaused && run.Status != models.ImportRunStatusPauseRequested {
		return
	}
	go o.handleImportRun(runID, run.ProposalID, "")
}

// CancelImport requests cancellation for an active import run.
func (o *ImportOrchestrator) CancelImport(runID uuid.UUID) bool {
	o.mu.Lock()
	cancel := o.activeCancels[runID]
	o.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

// PauseImport requests pause for an active import run (cooperative via ctx cancellation).
func (o *ImportOrchestrator) PauseImport(runID uuid.UUID) bool {
	if o.importRuns != nil {
		_ = o.importRuns.SetStatus(context.Background(), runID, models.ImportRunStatusPauseRequested)
	}
	return o.CancelImport(runID)
}

func (o *ImportOrchestrator) handleImportRun(runID uuid.UUID, proposalID uuid.UUID, cookieHeader string) {
	// Imports can be very long for large books (1000+ chapters) because parser-service uses real browser.
	// We allow a larger budget and rely on pause/cancel for control.
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	o.mu.Lock()
	o.activeCancels[runID] = cancel
	o.mu.Unlock()
	defer func() {
		cancel()
		o.mu.Lock()
		delete(o.activeCancels, runID)
		o.mu.Unlock()
	}()

	p, err := o.votingRepo.GetProposalByID(ctx, proposalID)
	if err != nil {
		o.logger.Error().Err(err).Str("proposal_id", proposalID.String()).Msg("Failed to load proposal for import")
		return
	}
	if p == nil {
		o.logger.Error().Str("proposal_id", proposalID.String()).Msg("Proposal not found for import")
		return
	}

	imp := o.pickImporter(p.OriginalLink)
	if imp == nil {
		o.logger.Error().
			Str("proposal_id", proposalID.String()).
			Str("original_link", p.OriginalLink).
			Msg("No importer registered for proposal original_link")
		return
	}

	o.logger.Info().
		Str("run_id", runID.String()).
		Str("proposal_id", proposalID.String()).
		Str("importer", imp.Name()).
		Msg("Starting import for daily vote winner")

	// Record run
	if o.importRuns != nil {
		// Create only if missing.
		if existing, err := o.importRuns.GetByID(ctx, runID); err != nil || existing == nil {
			_ = o.importRuns.Create(ctx, &models.ImportRun{
				ID:         runID,
				ProposalID: proposalID,
				Importer:   imp.Name(),
				Status:     models.ImportRunStatusRunning,
			})
		} else {
			_ = o.importRuns.SetStatus(ctx, runID, models.ImportRunStatusRunning)
		}
	}

	// Load checkpoint if any.
	var cp *importer.Checkpoint
	totalFromDB := 0
	if o.importRuns != nil {
		if run, err := o.importRuns.GetByID(ctx, runID); err == nil && run != nil {
			totalFromDB = run.ProgressTotal
			if len(run.Checkpoint) > 0 {
				var tmp importer.Checkpoint
				_ = json.Unmarshal(run.Checkpoint, &tmp)
				cp = &tmp
			}
		}
	}

	// Get cookies if not provided and run exists
	if cookieHeader == "" && o.cookiesRepo != nil {
		if cookie, err := o.cookiesRepo.GetByRunID(ctx, runID); err == nil && cookie != nil {
			cookieHeader = cookie.CookieHeader
			o.logger.Info().Str("run_id", runID.String()).Msg("Loaded cookies from database for import run")
		}
	}
	if cookieHeader != "" {
		o.logger.Info().Str("run_id", runID.String()).Int("cookie_len", len(cookieHeader)).Msg("Using cookies for import")
	} else {
		o.logger.Warn().Str("run_id", runID.String()).Msg("No cookies provided for import - may fail on Cloudflare")
	}

	onChapter := func(c *importer.Checkpoint, saved int) error {
		if o.importRuns == nil || c == nil {
			return nil
		}
		// Persist novel id early once known
		if c.NovelID != uuid.Nil {
			_ = o.importRuns.SetNovelID(ctx, runID, c.NovelID)
		}
		total := c.TotalChapters
		if total == 0 {
			total = totalFromDB
		}
		return o.importRuns.UpdateProgress(ctx, runID, c.NextIndex, total, c)
	}

	novelID, cp, err := imp.Import(ctx, o.db, p, o.uploadsDir, cp, onChapter, cookieHeader)
	if err != nil {
		errMsg := err.Error()
		if ctx.Err() == context.Canceled {
			// Decide: pause vs cancel, based on persisted status.
			if o.importRuns != nil {
				if run, e := o.importRuns.GetByID(context.Background(), runID); e == nil && run != nil && run.Status == models.ImportRunStatusPauseRequested {
					_ = o.importRuns.SetStatus(context.Background(), runID, models.ImportRunStatusPaused)
					o.logger.Warn().Str("run_id", runID.String()).Str("proposal_id", proposalID.String()).Msg("Import paused")
					return
				}
			}
			if o.importRuns != nil {
				cloudflareBlocked := isCloudflareError(errMsg)
				_ = o.importRuns.SetResult(context.Background(), runID, models.ImportRunStatusCancelled, nil, &errMsg, &cloudflareBlocked)
			}
			o.logger.Warn().Err(err).Str("run_id", runID.String()).Str("proposal_id", proposalID.String()).Msg("Import cancelled")
			return
		}
		if o.importRuns != nil {
			cloudflareBlocked := isCloudflareError(errMsg)
			_ = o.importRuns.SetResult(context.Background(), runID, models.ImportRunStatusFailed, nil, &errMsg, &cloudflareBlocked)
		}
		o.logger.Error().
			Err(err).
			Str("run_id", runID.String()).
			Str("proposal_id", proposalID.String()).
			Str("importer", imp.Name()).
			Msg("Import failed for daily vote winner")
		return
	}

	if err := o.votingRepo.SetProposalNovelID(ctx, proposalID, novelID); err != nil {
		errMsg := err.Error()
		if o.importRuns != nil {
			cloudflareBlocked := false
			_ = o.importRuns.SetResult(context.Background(), runID, models.ImportRunStatusFailed, &novelID, &errMsg, &cloudflareBlocked)
		}
		o.logger.Error().Err(err).Str("proposal_id", proposalID.String()).Str("novel_id", novelID.String()).Msg("Failed to link proposal->novel")
		return
	}

	o.logger.Info().
		Str("run_id", runID.String()).
		Str("proposal_id", proposalID.String()).
		Str("novel_id", novelID.String()).
		Msg("Proposal released into novel")

	if o.importRuns != nil {
		cloudflareBlocked := false
		_ = o.importRuns.SetResult(context.Background(), runID, models.ImportRunStatusSucceeded, &novelID, nil, &cloudflareBlocked)
	}

	if o.bus != nil {
		_ = o.bus.Publish(ctx, events.ProposalReleased{ProposalID: proposalID, NovelID: novelID})
	}
}

func (o *ImportOrchestrator) pickImporter(originalLink string) ProposalImporter {
	// normalize url (some users paste without scheme)
	if _, err := url.ParseRequestURI(originalLink); err != nil {
		// still allow CanImport() implementations to decide
	}
	for _, imp := range o.importers {
		if imp != nil && imp.CanImport(originalLink) {
			return imp
		}
	}
	return nil
}

