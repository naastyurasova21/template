package create

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/internal/domain/trip"
	"github.com/naastyurasova21/template/internal/model"
)

type Usecase struct {
	repo    TripRepository
	tx      TxManager
	metrics Metrics
	logger  Logger
}

func New(repo TripRepository, tx TxManager, metrics Metrics, logger Logger) *Usecase {
	return &Usecase{repo: repo, tx: tx, metrics: metrics, logger: logger}
}

type Input struct {
	UserID   uuid.UUID
	DriverID uuid.UUID
	StartLat float64
	StartLon float64
	EndLat   float64
	EndLon   float64
	Price    int64
}

func (u *Usecase) Execute(ctx context.Context, in Input) (*model.Trip, error) {
	now := time.Now().UTC()

	domainTrip, err := trip.New(
		in.UserID, in.DriverID,
		in.StartLat, in.StartLon, in.EndLat, in.EndLon,
		in.Price,
		now,
	)
	if err != nil {
		return nil, err
	}

	m := domainToModel(domainTrip)

	var saved *model.Trip
	err = u.tx.Do(ctx, func(ctx context.Context) error {
		var saveErr error
		saved, saveErr = u.repo.Save(ctx, m)
		return saveErr
	})
	if err != nil {
		return nil, err
	}

	u.metrics.IncCreated(ctx)
	u.logger.Info("trip created", "trip_id", saved.ID.String())
	return saved, nil
}

func domainToModel(t *trip.Trip) model.Trip {
	return model.Trip{
		ID:             t.ID,
		UserID:         t.UserID,
		DriverID:       t.DriverID,
		StartLatitude:  t.StartLatitude,
		StartLongitude: t.StartLongitude,
		EndLatitude:    t.EndLatitude,
		EndLongitude:   t.EndLongitude,
		Price:          t.Price,
		Status:         string(t.Status),
		StartedAt:      t.StartedAt,
		FinishedAt:     t.FinishedAt,
		LastPositionAt: t.LastPositionAt,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
