package models

import (
	"time"

	"github.com/google/uuid"
)

// XPEventType represents the type of XP-earning event
type XPEventType string

const (
	XPEventReadChapter XPEventType = "read_chapter"
	XPEventComment     XPEventType = "comment"
	XPEventDailyLogin  XPEventType = "daily_login"
	XPEventVote        XPEventType = "vote"
	XPEventProposal    XPEventType = "proposal"
	XPEventBookmark    XPEventType = "bookmark"
)

// XP rewards for each event type
var XPRewards = map[XPEventType]int64{
	XPEventReadChapter: 10,
	XPEventComment:     5,
	XPEventDailyLogin:  15,
	XPEventVote:        2,
	XPEventProposal:    50,
	XPEventBookmark:    3,
}

// UserXP represents a user's XP and level
type UserXP struct {
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	XPTotal   int64     `json:"xpTotal" db:"xp_total"`
	Level     int       `json:"level" db:"level"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// XPEvent represents an XP earning event
type XPEvent struct {
	ID        uuid.UUID   `json:"id" db:"id"`
	UserID    uuid.UUID   `json:"userId" db:"user_id"`
	Type      XPEventType `json:"type" db:"type"`
	Delta     int64       `json:"delta" db:"delta"`
	RefType   string      `json:"refType" db:"ref_type"`
	RefID     uuid.UUID   `json:"refId" db:"ref_id"`
	CreatedAt time.Time   `json:"createdAt" db:"created_at"`
}

// LevelInfo provides info about a level
type LevelInfo struct {
	Level         int   `json:"level"`
	XPRequired    int64 `json:"xpRequired"`
	XPForNextLevel int64 `json:"xpForNextLevel"`
	TotalXP       int64 `json:"totalXP"`
	CurrentXP     int64 `json:"currentXP"` // XP within current level
	Progress      float64 `json:"progress"` // 0.0 - 1.0
}

// Achievement represents an achievement/badge
type Achievement struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	IconKey     string    `json:"iconKey" db:"icon_key"`
	Condition   string    `json:"condition" db:"condition"` // JSON condition
	XPReward    int64     `json:"xpReward" db:"xp_reward"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// UserAchievement represents an unlocked achievement
type UserAchievement struct {
	UserID        uuid.UUID    `json:"userId" db:"user_id"`
	AchievementID uuid.UUID    `json:"achievementId" db:"achievement_id"`
	UnlockedAt    time.Time    `json:"unlockedAt" db:"unlocked_at"`
	Achievement   *Achievement `json:"achievement,omitempty"`
}

// UserStats represents aggregated user statistics
type UserStats struct {
	UserID           uuid.UUID `json:"userId" db:"user_id"`
	ChaptersRead     int       `json:"chaptersRead" db:"chapters_read"`
	ReadingTime      int64     `json:"readingTime" db:"reading_time"` // in seconds
	CommentsCount    int       `json:"commentsCount" db:"comments_count"`
	BookmarksCount   int       `json:"bookmarksCount" db:"bookmarks_count"`
	VotesSpent       int       `json:"votesSpent" db:"votes_spent"`
	ProposalsCreated int       `json:"proposalsCreated" db:"proposals_created"`
}

// XPLeaderboardEntry represents an entry in XP leaderboard
type XPLeaderboardEntry struct {
	Rank        int       `json:"rank"`
	UserID      uuid.UUID `json:"userId" db:"user_id"`
	DisplayName string    `json:"displayName" db:"display_name"`
	AvatarURL   *string   `json:"avatarUrl,omitempty" db:"avatar_url"`
	Level       int       `json:"level" db:"level"`
	XPTotal     int64     `json:"xpTotal" db:"xp_total"`
}

// CalculateLevel calculates level from total XP
// Level formula: level = floor(sqrt(xp / 100))
// XP for level n: xp = n^2 * 100
func CalculateLevel(xp int64) int {
	if xp < 100 {
		return 1
	}
	level := 1
	for {
		required := int64(level * level * 100)
		if xp < required {
			return level - 1
		}
		level++
		if level > 100 { // Max level cap
			return 100
		}
	}
}

// XPForLevel returns XP required to reach a level
func XPForLevel(level int) int64 {
	if level < 1 {
		return 0
	}
	return int64(level * level * 100)
}

// GetLevelInfo returns detailed level information
func GetLevelInfo(xpTotal int64) LevelInfo {
	level := CalculateLevel(xpTotal)
	xpRequired := XPForLevel(level)
	xpForNext := XPForLevel(level + 1)
	currentXP := xpTotal - xpRequired
	xpNeeded := xpForNext - xpRequired
	
	var progress float64
	if xpNeeded > 0 {
		progress = float64(currentXP) / float64(xpNeeded)
	}
	
	return LevelInfo{
		Level:          level,
		XPRequired:     xpRequired,
		XPForNextLevel: xpForNext,
		TotalXP:        xpTotal,
		CurrentXP:      currentXP,
		Progress:       progress,
	}
}
