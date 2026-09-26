package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/naastyurasova21/template/api"
)

func scanTrip(row pgx.Row) (*api.Trip, error) {
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

	return &api.Trip{
		Id:         id,
		UserId:     userId,
		DriverId:   driverId,
		StartPoint: api.Coordinates{Latitude: startLat, Longitude: startLng},
		EndPoint:   api.Coordinates{Latitude: endLat, Longitude: endLng},
		Price:      price,
		Status:     api.TripStatus(status),
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}, nil
}