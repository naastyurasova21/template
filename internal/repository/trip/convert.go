package trip

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/naastyurasova21/template/internal/model"
)

func scanTrip(row pgx.Row) (*model.Trip, error) {
	var (
		id         uuid.UUID
		userId     uuid.UUID
		driverId   uuid.UUID
		startLat   float64
		startLng   float64
		endLat     float64
		endLng     float64
		price      int64
		status     string
		startedAt  time.Time
		finishedAt *time.Time
	)

	err := row.Scan(
		&id, &userId, &driverId,
		&startLat, &startLng,
		&endLat, &endLng,
		&price, &status,
		&startedAt, &finishedAt,
	)
	if err != nil {
		return nil, err
	}

	return &model.Trip{
		ID:             id,
		UserID:         userId,
		DriverID:       driverId,
		StartLatitude:  startLat,
		StartLongitude: startLng,
		EndLatitude:    endLat,
		EndLongitude:   endLng,
		Price:          price,
		Status:         status,
		StartedAt:      startedAt,
		FinishedAt:     finishedAt,
	}, nil
}
