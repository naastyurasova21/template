package finish

import (
	"context"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/model"
)

type Usecase struct {
	repo    TripRepository
	tx      TxManager
	metrics Metrics
}

func New(repo TripRepository, tx TxManager, metrics Metrics) *Usecase {
	return &Usecase{repo: repo, tx: tx, metrics: metrics}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (*model.Trip, error) {
	var finished *model.Trip

	err := u.tx.Do(ctx, func(ctx context.Context) error {
		var err error
		finished, err = u.repo.Finish(ctx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	u.metrics.IncCompleted(ctx)

	return finished, nil
}
