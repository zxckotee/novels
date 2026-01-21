package handlers

import (
	"net/http"
	"os"

	"novels-backend/internal/jobs"
	"novels-backend/pkg/response"
	"github.com/rs/zerolog"
)

// JobsHandler exposes admin endpoints for running/checking background jobs.
type JobsHandler struct {
	scheduler *jobs.Scheduler
	logger    zerolog.Logger
	password  string
}

func NewJobsHandler(scheduler *jobs.Scheduler, logger zerolog.Logger) *JobsHandler {
	return &JobsHandler{
		scheduler: scheduler,
		logger:    logger.With().Str("handler", "jobs").Logger(),
		password:  os.Getenv("DAILY_VOTES_ENDPOINT_PASSWORD"),
	}
}

func (h *JobsHandler) authorizeDailyVotesEndpoint(w http.ResponseWriter, r *http.Request) bool {
	// If password is set, require it (header preferred; query allowed for quick testing).
	if h.password == "" {
		response.Error(w, http.StatusForbidden, "FORBIDDEN", "DAILY_VOTES_ENDPOINT_PASSWORD is not configured")
		return false
	}

	// Header: X-Daily-Votes-Password: <password>
	if r.Header.Get("X-Daily-Votes-Password") == h.password {
		return true
	}

	// Query: ?password=<password>
	if r.URL.Query().Get("password") == h.password {
		return true
	}

	response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid job password")
	return false
}

// GetDailyVotesStatus returns status of the last daily vote grant.
// GET /api/v1/admin/jobs/daily-votes/status
func (h *JobsHandler) GetDailyVotesStatus(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeDailyVotesEndpoint(w, r) {
		return
	}
	status, err := h.scheduler.GetDailyGrantStatus(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get daily vote grant status")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get job status")
		return
	}
	response.JSON(w, http.StatusOK, status)
}

// RunDailyVotesNow runs daily vote grant immediately.
// POST /api/v1/admin/jobs/daily-votes/run
func (h *JobsHandler) RunDailyVotesNow(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeDailyVotesEndpoint(w, r) {
		return
	}
	if err := h.scheduler.RunDailyVoteJobNow(r.Context()); err != nil {
		h.logger.Error().Err(err).Msg("Failed to run daily vote grant job now")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to run job")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "daily vote grant job executed"})
}

// GetWeeklyTicketsStatus returns status of the last weekly ticket grant.
// GET /api/v1/admin/jobs/weekly-tickets/status
func (h *JobsHandler) GetWeeklyTicketsStatus(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeDailyVotesEndpoint(w, r) {
		return
	}
	status, err := h.scheduler.GetWeeklyGrantStatus(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get weekly ticket grant status")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get job status")
		return
	}
	response.JSON(w, http.StatusOK, status)
}

// RunWeeklyTicketsNow runs weekly ticket grant immediately.
// POST /api/v1/admin/jobs/weekly-tickets/run
func (h *JobsHandler) RunWeeklyTicketsNow(w http.ResponseWriter, r *http.Request) {
	if !h.authorizeDailyVotesEndpoint(w, r) {
		return
	}
	if err := h.scheduler.RunWeeklyTicketJobNow(r.Context()); err != nil {
		h.logger.Error().Err(err).Msg("Failed to run weekly ticket grant job now")
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to run job")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "weekly ticket grant job executed"})
}

