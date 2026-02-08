package events

import "github.com/google/uuid"

const (
	EventDailyVoteWinnerSelected      = "daily_vote_winner_selected"
	EventTranslationVoteWinnerSelected = "translation_vote_winner_selected"
	EventProposalReleased             = "proposal_released"
)

type DailyVoteWinnerSelected struct {
	ProposalID uuid.UUID
}

func (DailyVoteWinnerSelected) Name() string { return EventDailyVoteWinnerSelected }

type TranslationVoteWinnerSelected struct {
	TargetID   uuid.UUID
	NovelID    *uuid.UUID
	ProposalID *uuid.UUID
}

func (TranslationVoteWinnerSelected) Name() string { return EventTranslationVoteWinnerSelected }

// ProposalReleased is fired when a proposal is successfully released into a novel (import done).
type ProposalReleased struct {
	ProposalID uuid.UUID
	NovelID    uuid.UUID
}

func (ProposalReleased) Name() string { return EventProposalReleased }

