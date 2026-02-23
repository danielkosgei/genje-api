package services

import (
	"context"

	"github.com/google/uuid"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type ElectionService struct {
	electionRepo *repository.ElectionRepo
}

func NewElectionService(er *repository.ElectionRepo) *ElectionService {
	return &ElectionService{electionRepo: er}
}

func (s *ElectionService) List(ctx context.Context) ([]models.Election, error) {
	return s.electionRepo.List(ctx)
}

func (s *ElectionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Election, error) {
	return s.electionRepo.GetByID(ctx, id)
}

func (s *ElectionService) GetCandidates(ctx context.Context, electionID uuid.UUID) ([]models.CandidacyDetail, error) {
	return s.electionRepo.GetCandidates(ctx, electionID)
}

func (s *ElectionService) GetResults(ctx context.Context, electionID uuid.UUID) ([]models.ResultSummary, error) {
	return s.electionRepo.GetResults(ctx, electionID)
}

func (s *ElectionService) GetTimeline(ctx context.Context, electionID uuid.UUID) ([]models.TimelineEvent, error) {
	return s.electionRepo.GetTimeline(ctx, electionID)
}
