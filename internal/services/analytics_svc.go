package services

import (
	"context"

	"jalada/internal/repository"
)

type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepo
	sentimentRepo *repository.SentimentRepo
}

func NewAnalyticsService(ar *repository.AnalyticsRepo, sr *repository.SentimentRepo) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: ar, sentimentRepo: sr}
}

func (s *AnalyticsService) GetPromiseAnalytics(ctx context.Context) (*repository.PromiseAnalytics, error) {
	return s.analyticsRepo.GetPromiseAnalytics(ctx)
}

func (s *AnalyticsService) GetIntegrityAnalytics(ctx context.Context) (*repository.IntegrityAnalytics, error) {
	return s.analyticsRepo.GetIntegrityAnalytics(ctx)
}

func (s *AnalyticsService) GetAttendanceAnalytics(ctx context.Context) (*repository.AttendanceAnalytics, error) {
	return s.analyticsRepo.GetAttendanceAnalytics(ctx)
}

func (s *AnalyticsService) GetTrending(ctx context.Context, limit int) ([]repository.TrendingItem, error) {
	return s.analyticsRepo.GetTrending(ctx, limit)
}

func (s *AnalyticsService) GetNationalSentiment(ctx context.Context) (interface{}, error) {
	return s.sentimentRepo.GetNationalSentiment(ctx)
}
