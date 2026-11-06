package get

import (
	"context"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/model"
)

type Usecase struct {
	repo TripRepository
}

func New(repo TripRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*model.Trip, error) {
	return u.repo.GetByID(ctx, id)
}
