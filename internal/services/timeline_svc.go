package services

import (
	"context"

	"jalada/internal/models"
	"jalada/internal/repository"
)

type TimelineService struct {
	eventRepo *repository.EventRepo
}

func NewTimelineService(er *repository.EventRepo) *TimelineService {
	return &TimelineService{eventRepo: er}
}

func (s *TimelineService) GetElectionTimeline(ctx context.Context) ([]models.TimelineEvent, error) {
	return s.eventRepo.GetTimelineEvents(ctx)
}

func (s *TimelineService) ListEvents(ctx context.Context, limit, offset int) ([]models.Event, int, error) {
	return s.eventRepo.ListEvents(ctx, limit, offset)
}
