package trip

import (
	"context"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/model"
)

type Repository interface {
	Save(ctx context.Context, t model.Trip) (*model.Trip, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Trip, error)
	Finish(ctx context.Context, id uuid.UUID) (*model.Trip, error)
}
