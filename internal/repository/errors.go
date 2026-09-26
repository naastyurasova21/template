package repository

import "errors"


var (
	ErrTripNotFound         = errors.New("trip not found")
	ErrDriverBusy           = errors.New("driver already has an active trip")
	ErrTripAlreadyCompleted = errors.New("trip already completed")
)