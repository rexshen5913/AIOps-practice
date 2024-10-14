package service

import (
	"context"

	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/domain"
	"github.com/rexshen5913/AIOps-pracgice/Week5/GetCrd/internal/repository"
)

type AIOpsService struct {
	repo *repository.AIOpsRepository
}

func NewAIOpsService(repo *repository.AIOpsRepository) *AIOpsService {
	return &AIOpsService{
		repo: repo,
	}
}

func (s *AIOpsService) ListAIOpsResource(ctx context.Context, kind string) ([]domain.AIOps, error) {
	return s.repo.ListAIOpsResource(ctx, kind)
}
