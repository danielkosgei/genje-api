package services

import (
	"context"

	"github.com/google/uuid"

	"jalada/internal/models"
)

// GetProfile returns the sourced figures for one politician's dossier.
func (s *PoliticianService) GetProfile(ctx context.Context, politicianID uuid.UUID, slug string) (models.PoliticianProfile, error) {
	return s.politicianRepo.GetProfile(ctx, politicianID, slug)
}
