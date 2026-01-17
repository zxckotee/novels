package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AppSetting represents a system setting
type AppSetting struct {
	Key         string          `json:"key" db:"key"`
	Value       json.RawMessage `json:"value" db:"value"`
	Description string          `json:"description,omitempty" db:"description"`
	UpdatedBy   *uuid.UUID      `json:"updatedBy,omitempty" db:"updated_by"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
}

// AdminAuditLog represents an admin action log entry
type AdminAuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	ActorUserID  uuid.UUID       `json:"actorUserId" db:"actor_user_id"`
	Action       string          `json:"action" db:"action"`
	EntityType   string          `json:"entityType" db:"entity_type"`
	EntityID     *uuid.UUID      `json:"entityId,omitempty" db:"entity_id"`
	Details      json.RawMessage `json:"details,omitempty" db:"details"`
	IPAddress    string          `json:"ipAddress,omitempty" db:"ip_address"`
	UserAgent    string          `json:"userAgent,omitempty" db:"user_agent"`
	CreatedAt    time.Time       `json:"createdAt" db:"created_at"`

	// Populated from joins
	ActorUser *User `json:"actorUser,omitempty"`
}

// UpdateSettingRequest represents the request to update a setting
type UpdateSettingRequest struct {
	Value json.RawMessage `json:"value" validate:"required"`
}

// AdminLogsFilter represents filters for listing admin logs
type AdminLogsFilter struct {
	ActorUserID *uuid.UUID `json:"actorUserId,omitempty"`
	Action      string     `json:"action,omitempty"`
	EntityType  string     `json:"entityType,omitempty"`
	EntityID    *uuid.UUID `json:"entityId,omitempty"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	Page        int        `json:"page"`
	Limit       int        `json:"limit"`
}

// AdminLogsResponse represents a paginated list of admin logs
type AdminLogsResponse struct {
	Logs       []AdminAuditLog `json:"logs"`
	TotalCount int             `json:"totalCount"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

// AdminStatsOverview represents overview statistics for the admin panel
type AdminStatsOverview struct {
	TotalNovels        int     `json:"totalNovels"`
	TotalChapters      int     `json:"totalChapters"`
	TotalUsers         int     `json:"totalUsers"`
	TotalComments      int     `json:"totalComments"`
	PendingReports     int     `json:"pendingReports"`
	NewUsersThisWeek   int     `json:"newUsersThisWeek"`
	NewNovelsThisWeek  int     `json:"newNovelsThisWeek"`
	NewChaptersThisWeek int    `json:"newChaptersThisWeek"`
	AvgChaptersPerDay  float64 `json:"avgChaptersPerDay"`
	AvgCommentsPerDay  float64 `json:"avgCommentsPerDay"`
}

// UserManagementRequest represents requests for user management
type BanUserRequest struct {
	Reason string `json:"reason" validate:"required,min=10,max=500"`
}

type UpdateUserRolesRequest struct {
	Roles []string `json:"roles" validate:"required,dive,oneof=user premium moderator admin"`
}

// UsersFilter represents filters for listing users
type UsersFilter struct {
	Query   string `json:"query,omitempty"`
	Role    string `json:"role,omitempty"`
	Banned  *bool  `json:"banned,omitempty"`
	Sort    string `json:"sort,omitempty"`  // created, login, name
	Order   string `json:"order,omitempty"` // asc, desc
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

// UsersResponse represents a paginated list of users
type UsersResponse struct {
	Users      []User `json:"users"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

// AdminCommentsFilter represents filters for admin comment listing
type AdminCommentsFilter struct {
	TargetType TargetType `json:"targetType,omitempty"`
	TargetID   *uuid.UUID `json:"targetId,omitempty"`
	UserID     *uuid.UUID `json:"userId,omitempty"`
	IsDeleted  *bool      `json:"isDeleted,omitempty"`
	Sort       string     `json:"sort,omitempty"` // newest, oldest, reports
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

// ReportsFilter represents filters for listing comment reports
type ReportsFilter struct {
	Status     string     `json:"status,omitempty"` // pending, resolved, dismissed
	CommentID  *uuid.UUID `json:"commentId,omitempty"`
	ReporterID *uuid.UUID `json:"reporterId,omitempty"`
	Sort       string     `json:"sort,omitempty"` // newest, oldest
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
}

// ReportsResponse represents a paginated list of reports
type ReportsResponse struct {
	Reports    []CommentReport `json:"reports"`
	TotalCount int             `json:"totalCount"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
}

// ResolveReportRequest represents the request to resolve a report
type ResolveReportRequest struct {
	Action string `json:"action" validate:"required,oneof=resolve dismiss delete_comment"`
	Reason string `json:"reason,omitempty" validate:"max=500"`
}
