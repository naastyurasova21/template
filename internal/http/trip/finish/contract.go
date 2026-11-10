package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/model"
)

type Usecase interface {
	Execute(ctx context.Context, id uuid.UUID) (*model.Trip, error)
}
