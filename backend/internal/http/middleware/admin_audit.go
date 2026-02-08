package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AdminAuditLogger is implemented by service.AdminService (via LogAction).
type AdminAuditLogger interface {
	LogAction(ctx context.Context, actorUserID uuid.UUID, action, entityType string, entityID *uuid.UUID, details json.RawMessage, ipAddress, userAgent string) error
}

// NOTE: We intentionally log only mutating requests by default to avoid flooding the audit table.
// This is audit-log of admin actions, not server runtime logs.
func AdminAuditMutations(audit AdminAuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := &statusWriter{ResponseWriter: w, status: 200}
			start := time.Now()
			next.ServeHTTP(ww, r)

			// Log only mutating methods.
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			default:
				return
			}
			if audit == nil {
				return
			}

			uidStr := GetUserID(r.Context())
			actorID, err := uuid.Parse(uidStr)
			if err != nil || actorID == uuid.Nil {
				return
			}

			routePattern := r.URL.Path
			if rc := chi.RouteContext(r.Context()); rc != nil {
				if p := rc.RoutePattern(); p != "" {
					routePattern = p
				}
			}

			detailsObj := map[string]any{
				"method":   r.Method,
				"path":     r.URL.Path,
				"pattern":  routePattern,
				"query":    r.URL.RawQuery,
				"status":   ww.status,
				"duration": time.Since(start).String(),
			}
			detailsJSON, _ := json.Marshal(detailsObj)

			ip := r.RemoteAddr
			if host, _, e := net.SplitHostPort(r.RemoteAddr); e == nil {
				ip = host
			}
			_ = audit.LogAction(r.Context(), actorID, fmt.Sprintf("%s %s", r.Method, routePattern), "http", nil, detailsJSON, ip, r.UserAgent())
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

