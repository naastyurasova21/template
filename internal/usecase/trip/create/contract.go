//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

package create

import (
	"context"

	"github.com/naastyurasova21/template/internal/model"
)

type (
	TripRepository interface {
		Save(ctx context.Context, t model.Trip) (*model.Trip, error)
	}

	TxManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
	
	Metrics interface {
	IncCreated(ctx context.Context)
    }
)

