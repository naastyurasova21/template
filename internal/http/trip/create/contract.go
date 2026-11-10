package create

import (
	"context"

	"github.com/naastyurasova21/template/internal/model"
	"github.com/naastyurasova21/template/internal/usecase/trip/create"
)

type Usecase interface {
	Execute(ctx context.Context, in create.Input) (*model.Trip, error)
}
