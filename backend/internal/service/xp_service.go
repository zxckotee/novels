package service

import (
	"context"

	"github.com/google/uuid"
	"novels-backend/internal/domain/models"
	"novels-backend/internal/repository"
)

type XPService struct {
	xpRepo *repository.XPRepository
}

func NewXPService(xpRepo *repository.XPRepository) *XPService {
	return &XPService{xpRepo: xpRepo}
}

// GetUserXP retrieves user's XP and level info
func (s *XPService) GetUserXP(ctx context.Context, userID uuid.UUID) (*models.UserXP, *models.LevelInfo, error) {
	xp, err := s.xpRepo.GetUserXP(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	
	levelInfo := models.GetLevelInfo(xp.XPTotal)
	return xp, &levelInfo, nil
}

// AwardXP awards XP for an action (with idempotency check)
func (s *XPService) AwardXP(ctx context.Context, userID uuid.UUID, eventType models.XPEventType, amount int64, refType string, refID uuid.UUID) error {
	// Check if event already exists (idempotency)
	exists, err := s.xpRepo.CheckEventExists(ctx, userID, eventType, refType, refID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already awarded
	}
	
	// Use default amount if not specified
	if amount == 0 {
		amount = models.XPRewards[eventType]
	}
	
	// Record event
	event := &models.XPEvent{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    eventType,
		Delta:   amount,
		RefType: refType,
		RefID:   refID,
	}
	
	err = s.xpRepo.AddXPEvent(ctx, event)
	if err != nil {
		return err
	}
	
	// Update total XP
	_, err = s.xpRepo.CreateOrUpdateXP(ctx, userID, amount)
	if err != nil {
		return err
	}
	
	// Check for achievements (async in production)
	go s.checkAchievements(context.Background(), userID)
	
	return nil
}

// GetXPEvents retrieves XP events for a user
func (s *XPService) GetXPEvents(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.XPEvent, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit
	
	return s.xpRepo.GetXPEvents(ctx, userID, limit, offset)
}

// GetLeaderboard retrieves XP leaderboard
func (s *XPService) GetLeaderboard(ctx context.Context, limit int) ([]models.XPLeaderboardEntry, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.xpRepo.GetLeaderboard(ctx, limit)
}

// GetUserRank retrieves user's rank in leaderboard
func (s *XPService) GetUserRank(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.xpRepo.GetUserRank(ctx, userID)
}

// GetUserStats retrieves aggregated user statistics
func (s *XPService) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.UserStats, error) {
	return s.xpRepo.GetUserStats(ctx, userID)
}

// GetAchievements retrieves all achievements
func (s *XPService) GetAchievements(ctx context.Context) ([]models.Achievement, error) {
	return s.xpRepo.GetAchievements(ctx)
}

// GetUserAchievements retrieves user's unlocked achievements
func (s *XPService) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]models.UserAchievement, error) {
	return s.xpRepo.GetUserAchievements(ctx, userID)
}

// checkAchievements checks and unlocks achievements for user
func (s *XPService) checkAchievements(ctx context.Context, userID uuid.UUID) {
	stats, err := s.xpRepo.GetUserStats(ctx, userID)
	if err != nil {
		return
	}
	
	achievements, err := s.xpRepo.GetAchievements(ctx)
	if err != nil {
		return
	}
	
	for _, achievement := range achievements {
		// Check if already unlocked
		hasIt, _ := s.xpRepo.HasAchievement(ctx, userID, achievement.ID)
		if hasIt {
			continue
		}
		
		// Check conditions based on achievement code
		shouldUnlock := false
		switch achievement.Code {
		case "first_chapter":
			shouldUnlock = stats.ChaptersRead >= 1
		case "bookworm_10":
			shouldUnlock = stats.ChaptersRead >= 10
		case "bookworm_100":
			shouldUnlock = stats.ChaptersRead >= 100
		case "bookworm_1000":
			shouldUnlock = stats.ChaptersRead >= 1000
		case "first_comment":
			shouldUnlock = stats.CommentsCount >= 1
		case "commentator_10":
			shouldUnlock = stats.CommentsCount >= 10
		case "commentator_100":
			shouldUnlock = stats.CommentsCount >= 100
		case "first_bookmark":
			shouldUnlock = stats.BookmarksCount >= 1
		case "collector_10":
			shouldUnlock = stats.BookmarksCount >= 10
		case "collector_100":
			shouldUnlock = stats.BookmarksCount >= 100
		}
		
		if shouldUnlock {
			_ = s.xpRepo.UnlockAchievement(ctx, userID, achievement.ID)
			// Award XP for achievement
			if achievement.XPReward > 0 {
				_ = s.AwardXP(ctx, userID, models.XPEventType("achievement"), achievement.XPReward, "achievement", achievement.ID)
			}
		}
	}
}

// InitializeUserXP creates initial XP record for new user
func (s *XPService) InitializeUserXP(ctx context.Context, userID uuid.UUID) error {
	_, err := s.xpRepo.CreateOrUpdateXP(ctx, userID, 0)
	return err
}
