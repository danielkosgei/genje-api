package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type PoliticianService struct {
	politicianRepo *repository.PoliticianRepo
	newsRepo       *repository.NewsRepo
	sentimentRepo  *repository.SentimentRepo
	eventRepo      *repository.EventRepo
}

func NewPoliticianService(
	pr *repository.PoliticianRepo,
	nr *repository.NewsRepo,
	sr *repository.SentimentRepo,
	er *repository.EventRepo,
) *PoliticianService {
	return &PoliticianService{
		politicianRepo: pr,
		newsRepo:       nr,
		sentimentRepo:  sr,
		eventRepo:      er,
	}
}

func (s *PoliticianService) List(ctx context.Context, f models.PoliticianFilter) ([]models.PoliticianSummary, int, error) {
	return s.politicianRepo.List(ctx, f)
}

func (s *PoliticianService) GetDossier(ctx context.Context, slug string) (*models.PoliticianDossier, error) {
	p, err := s.politicianRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get politician: %w", err)
	}
	if p == nil {
		return nil, nil
	}

	dossier := &models.PoliticianDossier{Politician: *p}

	history, err := s.politicianRepo.GetPartyHistory(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("get party history: %w", err)
	}
	dossier.PartyHistory = history
	for i, m := range history {
		if m.LeftDate == nil {
			dossier.CurrentParty = &history[i]
			break
		}
	}

	candidacies, err := s.politicianRepo.GetCandidacies(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("get candidacies: %w", err)
	}
	dossier.Candidacies = candidacies

	flags, err := s.politicianRepo.GetIntegrityFlags(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("get integrity flags: %w", err)
	}
	dossier.IntegrityFlags = flags

	return dossier, nil
}

func (s *PoliticianService) GetBySlug(ctx context.Context, slug string) (*models.Politician, error) {
	return s.politicianRepo.GetBySlug(ctx, slug)
}

func (s *PoliticianService) GetCourtCases(ctx context.Context, politicianID uuid.UUID) ([]models.CourtCase, error) {
	return s.politicianRepo.GetCourtCases(ctx, politicianID)
}

func (s *PoliticianService) GetPromises(ctx context.Context, politicianID uuid.UUID) ([]models.Promise, error) {
	return s.politicianRepo.GetPromises(ctx, politicianID)
}

func (s *PoliticianService) GetPromiseStats(ctx context.Context, politicianID uuid.UUID) (*models.PromiseStats, error) {
	return s.politicianRepo.GetPromiseStats(ctx, politicianID)
}

func (s *PoliticianService) GetAchievements(ctx context.Context, politicianID uuid.UUID) ([]models.Achievement, error) {
	return s.politicianRepo.GetAchievements(ctx, politicianID)
}

func (s *PoliticianService) GetControversies(ctx context.Context, politicianID uuid.UUID) ([]models.Controversy, error) {
	return s.politicianRepo.GetControversies(ctx, politicianID)
}

func (s *PoliticianService) GetAffiliations(ctx context.Context, politicianID uuid.UUID) ([]models.AffiliationDetail, error) {
	return s.politicianRepo.GetAffiliations(ctx, politicianID)
}

func (s *PoliticianService) GetAssetDeclarations(ctx context.Context, politicianID uuid.UUID) ([]models.AssetDeclaration, error) {
	return s.politicianRepo.GetAssetDeclarations(ctx, politicianID)
}

func (s *PoliticianService) GetVotingRecords(ctx context.Context, politicianID uuid.UUID) ([]models.VotingRecord, error) {
	return s.politicianRepo.GetVotingRecords(ctx, politicianID)
}

func (s *PoliticianService) GetAttendanceStats(ctx context.Context, politicianID uuid.UUID) (*models.AttendanceStats, error) {
	return s.politicianRepo.GetAttendanceStats(ctx, politicianID)
}

func (s *PoliticianService) GetNews(ctx context.Context, politicianID uuid.UUID, limit, offset int) ([]models.NewsArticle, int, error) {
	return s.newsRepo.GetArticlesByPolitician(ctx, politicianID, limit, offset)
}

func (s *PoliticianService) GetSentiment(ctx context.Context, politicianID uuid.UUID, limit int) ([]models.SentimentSnapshot, error) {
	return s.sentimentRepo.GetByPolitician(ctx, politicianID, limit)
}

func (s *PoliticianService) GetEvents(ctx context.Context, politicianID uuid.UUID, limit, offset int) ([]models.Event, int, error) {
	return s.eventRepo.GetEventsByPolitician(ctx, politicianID, limit, offset)
}
