package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"novels-backend/internal/domain/models"
	"novels-backend/internal/jobs"
	"novels-backend/internal/orchestrator"
	"novels-backend/internal/repository"
	"novels-backend/pkg/response"
)

// OpsHandler is an admin-only handler for operational controls: manual job triggers, import runs, etc.
type OpsHandler struct {
	scheduler        *jobs.Scheduler
	orchestrator     *orchestrator.ImportOrchestrator
	importRunsRepo   *repository.ImportRunsRepository
	cookiesRepo      *repository.ImportRunCookiesRepository
	translationRepo  *repository.TranslationVotingRepository
	votingRepo       *repository.VotingRepository
	logger           zerolog.Logger
}

func NewOpsHandler(
	scheduler *jobs.Scheduler,
	orchestrator *orchestrator.ImportOrchestrator,
	importRunsRepo *repository.ImportRunsRepository,
	cookiesRepo *repository.ImportRunCookiesRepository,
	translationRepo *repository.TranslationVotingRepository,
	votingRepo *repository.VotingRepository,
	logger zerolog.Logger,
) *OpsHandler {
	return &OpsHandler{
		scheduler:       scheduler,
		orchestrator:    orchestrator,
		importRunsRepo:  importRunsRepo,
		cookiesRepo:     cookiesRepo,
		translationRepo: translationRepo,
		votingRepo:      votingRepo,
		logger:          logger.With().Str("handler", "ops").Logger(),
	}
}

// POST /api/v1/admin/ops/jobs/voting-winner/run
func (h *OpsHandler) RunVotingWinnerNow(w http.ResponseWriter, r *http.Request) {
	if h.scheduler == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "scheduler is not configured")
		return
	}
	force := r.URL.Query().Get("force") == "1" || r.URL.Query().Get("force") == "true"
	var err error
	if force {
		err = h.scheduler.RunVotingWinnerJobNowForce(r.Context())
	} else {
		err = h.scheduler.RunVotingWinnerJobNow(r.Context())
	}
	if err != nil {
		h.logger.Error().Err(err).Msg("RunVotingWinnerJobNow failed")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to run voting winner job")
		return
	}
	if force {
		response.OK(w, map[string]string{"message": "voting winner job executed (force)"})
		return
	}
	response.OK(w, map[string]string{"message": "voting winner job executed"})
}

// POST /api/v1/admin/ops/jobs/translation-winner/run
func (h *OpsHandler) RunTranslationWinnerNow(w http.ResponseWriter, r *http.Request) {
	if h.scheduler == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "scheduler is not configured")
		return
	}
	force := r.URL.Query().Get("force") == "1" || r.URL.Query().Get("force") == "true"
	var err error
	if force {
		err = h.scheduler.RunTranslationWinnerJobNowForce(r.Context())
	} else {
		err = h.scheduler.RunTranslationWinnerJobNow(r.Context())
	}
	if err != nil {
		h.logger.Error().Err(err).Msg("RunTranslationWinnerJobNow failed")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to run translation winner job")
		return
	}
	if force {
		response.OK(w, map[string]string{"message": "translation winner job executed (force)"})
		return
	}
	response.OK(w, map[string]string{"message": "translation winner job executed"})
}

// GET /api/v1/admin/ops/import-runs?limit=50&status=running
func (h *OpsHandler) ListImportRuns(w http.ResponseWriter, r *http.Request) {
	if h.importRunsRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "import runs repo is not configured")
		return
	}
	limit := parseIntQuery(r, "limit", 50)
	var st *models.ImportRunStatus
	if s := r.URL.Query().Get("status"); s != "" {
		tmp := models.ImportRunStatus(s)
		st = &tmp
	}
	runs, err := h.importRunsRepo.List(r.Context(), limit, st)
	if err != nil {
		h.logger.Error().Err(err).Msg("List import runs failed")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list import runs")
		return
	}
	response.OK(w, map[string]any{"runs": runs})
}

// POST /api/v1/admin/ops/import-runs/{id}/cancel
func (h *OpsHandler) CancelImportRun(w http.ResponseWriter, r *http.Request) {
	if h.orchestrator == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "orchestrator is not configured")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid run id")
		return
	}
	ok := h.orchestrator.CancelImport(runID)
	if !ok {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "run is not active")
		return
	}
	response.OK(w, map[string]string{"message": "cancel requested"})
}

// POST /api/v1/admin/ops/import-runs/{id}/pause
func (h *OpsHandler) PauseImportRun(w http.ResponseWriter, r *http.Request) {
	if h.orchestrator == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "orchestrator is not configured")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid run id")
		return
	}
	ok := h.orchestrator.PauseImport(runID)
	if !ok {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "run is not active")
		return
	}
	response.OK(w, map[string]string{"message": "pause requested"})
}

// POST /api/v1/admin/ops/import-runs/{id}/resume
func (h *OpsHandler) ResumeImportRun(w http.ResponseWriter, r *http.Request) {
	if h.orchestrator == nil || h.importRunsRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "orchestrator/import repo is not configured")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid run id")
		return
	}
	run, err := h.importRunsRepo.GetByID(r.Context(), runID)
	if err != nil || run == nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "run not found")
		return
	}
	// Only resume paused runs
	if run.Status != models.ImportRunStatusPaused && run.Status != models.ImportRunStatusPauseRequested {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "run is not paused")
		return
	}
	h.orchestrator.ResumeImportAsync(runID)
	response.OK(w, map[string]string{"message": "resume requested"})
}

type runImportReq struct {
	ProposalID string `json:"proposalId"`
}

// POST /api/v1/admin/ops/imports/run
func (h *OpsHandler) RunImportNow(w http.ResponseWriter, r *http.Request) {
	if h.orchestrator == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "orchestrator is not configured")
		return
	}
	var req runImportReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("RunImportNow: failed to decode request body")
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.ProposalID == "" {
		h.logger.Warn().Msg("RunImportNow: proposalId is empty")
		response.BadRequest(w, "proposalId is required")
		return
	}
	
	// Try to parse as UUID first
	pid, err := uuid.Parse(req.ProposalID)
	if err != nil {
		// If it's not a valid UUID, try to find by short ID (first 8 chars)
		if len(req.ProposalID) == 8 {
			// Search for proposal by short ID prefix
			if h.votingRepo == nil {
				h.logger.Error().Msg("RunImportNow: voting repo is not configured")
				response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "voting repo is not configured")
				return
			}
			foundID, err := h.votingRepo.GetProposalByShortID(r.Context(), req.ProposalID)
			if err != nil {
				h.logger.Error().Str("proposalId", req.ProposalID).Err(err).Msg("RunImportNow: failed to search proposal by short ID")
				response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search proposal")
				return
			}
			if foundID == nil {
				h.logger.Warn().Str("proposalId", req.ProposalID).Msg("RunImportNow: proposal not found by short ID")
				response.BadRequest(w, fmt.Sprintf("proposal not found: %s", req.ProposalID))
				return
			}
			pid = *foundID
			h.logger.Info().Str("shortId", req.ProposalID).Str("fullId", pid.String()).Msg("RunImportNow: found proposal by short ID")
		} else {
			h.logger.Warn().Str("proposalId", req.ProposalID).Err(err).Msg("RunImportNow: invalid proposalId format")
			response.BadRequest(w, fmt.Sprintf("invalid proposalId: %v", err))
			return
		}
	}
	runID := h.orchestrator.StartImportAsync(pid)
	h.logger.Info().Str("proposalId", pid.String()).Str("runId", runID.String()).Msg("RunImportNow: import started")
	response.OK(w, map[string]string{"runId": runID.String()})
}

// GET /api/v1/admin/ops/translation-targets?limit=100
func (h *OpsHandler) ListTranslationTargets(w http.ResponseWriter, r *http.Request) {
	if h.translationRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "translation repo is not configured")
		return
	}
	limit := parseIntQuery(r, "limit", 100)
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	entries, err := h.translationRepo.ListTargetsForOps(r.Context(), limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("ListTargetsForOps failed")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list translation targets")
		return
	}
	response.OK(w, map[string]any{"entries": entries})
}

type setTranslationTargetStatusReq struct {
	Status string `json:"status"`
}

// POST /api/v1/admin/ops/translation-targets/{id}/status
func (h *OpsHandler) SetTranslationTargetStatus(w http.ResponseWriter, r *http.Request) {
	if h.translationRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "translation repo is not configured")
		return
	}
	idStr := chi.URLParam(r, "id")
	tid, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid target id")
		return
	}
	var req setTranslationTargetStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	status := req.Status
	if status == "" {
		response.BadRequest(w, "status is required")
		return
	}
	// Accept only known statuses (string check; repository uses typed enum but it's just string).
	switch status {
	case "voting", "waiting_release", "translating", "completed", "cancelled":
	default:
		response.BadRequest(w, "invalid status")
		return
	}
	if err := h.translationRepo.UpdateTargetStatus(r.Context(), tid, models.TranslationVoteTargetStatus(status)); err != nil {
		h.logger.Error().Err(err).Msg("UpdateTargetStatus failed")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update status")
		return
	}
	response.OK(w, map[string]string{"message": "status updated"})
}

// GET /api/v1/admin/ops/import-runs/{id}/cookies
func (h *OpsHandler) GetImportRunCookies(w http.ResponseWriter, r *http.Request) {
	if h.cookiesRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "cookies repo is not configured")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid run id")
		return
	}
	cookie, err := h.cookiesRepo.GetByRunID(r.Context(), runID)
	if err != nil {
		// Not found is OK - return null
		response.OK(w, map[string]interface{}{"cookie": nil})
		return
	}
	response.OK(w, map[string]interface{}{"cookie": cookie})
}

type updateCookiesReq struct {
	CookieHeader string `json:"cookieHeader"`
}

// PUT /api/v1/admin/ops/import-runs/{id}/cookies
func (h *OpsHandler) UpdateImportRunCookies(w http.ResponseWriter, r *http.Request) {
	if h.cookiesRepo == nil || h.importRunsRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "cookies/import repo is not configured")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid run id")
		return
	}
	// Verify run exists
	run, err := h.importRunsRepo.GetByID(r.Context(), runID)
	if err != nil || run == nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "run not found")
		return
	}
	var req updateCookiesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if strings.TrimSpace(req.CookieHeader) == "" {
		response.BadRequest(w, "cookieHeader is required")
		return
	}
	if err := h.cookiesRepo.Upsert(r.Context(), runID, strings.TrimSpace(req.CookieHeader)); err != nil {
		h.logger.Error().Err(err).Str("run_id", runID.String()).Msg("Failed to upsert cookies")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save cookies")
		return
	}
	response.OK(w, map[string]string{"message": "cookies saved"})
}

// POST /api/v1/admin/ops/import-runs/{id}/retry
func (h *OpsHandler) RetryImportRun(w http.ResponseWriter, r *http.Request) {
	if h.orchestrator == nil || h.importRunsRepo == nil || h.cookiesRepo == nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "orchestrator/import/cookies repo is not configured")
		return
	}
	runID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid run id")
		return
	}
	// Get the failed run
	run, err := h.importRunsRepo.GetByID(r.Context(), runID)
	if err != nil || run == nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "run not found")
		return
	}
	// Only retry failed runs that were blocked by Cloudflare
	if run.Status != models.ImportRunStatusFailed {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "run is not failed")
		return
	}
	if !run.CloudflareBlocked {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "run was not blocked by Cloudflare")
		return
	}
	// Get cookies for this run
	cookie, err := h.cookiesRepo.GetByRunID(r.Context(), runID)
	if err != nil || cookie == nil {
		response.Error(w, http.StatusBadRequest, "BAD_REQUEST", "no cookies found for this run. Please set cookies first")
		return
	}
	// Start a new import run with the same proposal
	newRunID := h.orchestrator.StartImportAsyncWithCookies(run.ProposalID, cookie.CookieHeader)
	response.OK(w, map[string]string{"runId": newRunID.String(), "message": "retry started"})
}
