package trip

import (
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DriverID       uuid.UUID
	StartLatitude  float64
	StartLongitude float64
	EndLatitude    float64
	EndLongitude   float64
	Price          int64
	Status         Status
	StartedAt      time.Time
	FinishedAt     *time.Time
	LastPositionAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func New(
	userID, driverID uuid.UUID,
	startLat, startLon, endLat, endLon float64,
	price int64,
	now time.Time,
) (*Trip, error) {
	t := &Trip{
		ID:             uuid.New(),
		UserID:         userID,
		DriverID:       driverID,
		StartLatitude:  startLat,
		StartLongitude: startLon,
		EndLatitude:    endLat,
		EndLongitude:   endLon,
		Price:          price,
		Status:         StatusActive,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Trip) Finish(at time.Time) error {
	if t.Status != StatusActive {
		return ErrTripCompleted
	}

	t.Status = StatusCompleted
	t.FinishedAt = &at
	t.UpdatedAt = at

	return nil
}

func (t *Trip) Validate() error {
	if t.UserID == uuid.Nil {
		return newValidationError("user_id", "must not be empty")
	}
	if t.DriverID == uuid.Nil {
		return newValidationError("driver_id", "must not be empty")
	}

	if t.Price < 0 {
		return newValidationError("price", "must be greater than or equal to 0")
	}

	if t.StartLatitude < -90 || t.StartLatitude > 90 {
		return newValidationError("start_point.latitude", "must be in range [-90, 90]")
	}
	if t.StartLongitude < -180 || t.StartLongitude > 180 {
		return newValidationError("start_point.longitude", "must be in range [-180, 180]")
	}

	if t.EndLatitude < -90 || t.EndLatitude > 90 {
		return newValidationError("end_point.latitude", "must be in range [-90, 90]")
	}
	if t.EndLongitude < -180 || t.EndLongitude > 180 {
		return newValidationError("end_point.longitude", "must be in range [-180, 180]")
	}

	return nil
}
