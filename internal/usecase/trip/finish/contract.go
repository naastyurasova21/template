//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/model"
)

type TripRepository interface {
	Finish(ctx context.Context, id uuid.UUID) (*model.Trip, error)
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
