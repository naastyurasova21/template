package domainerror

import (
	"errors"
	"net/http"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/domain/trip"
)

func Problem(err error) (int, api.Problem) {
	switch {
	case errors.Is(err, trip.ErrTripNotFound):
		return http.StatusNotFound, newProblem(http.StatusNotFound, "trip_not_found", "Trip not found")
	case errors.Is(err, trip.ErrTripCompleted):
		return http.StatusConflict, newProblem(http.StatusConflict, "trip_completed", "Trip already completed")
	case errors.Is(err, trip.ErrDriverBusy):
		return http.StatusConflict, newProblem(http.StatusConflict, "driver_busy", "Driver already has an active trip")
	default:
		var verr *trip.ValidationError
		if errors.As(err, &verr) {
			return http.StatusBadRequest, newProblem(http.StatusBadRequest, "invalid_request", verr.Error())
		}
		return http.StatusInternalServerError, newProblem(http.StatusInternalServerError, "internal_error", "Internal error")
	}
}

func newProblem(status int, code, detail string) api.Problem {
	title := "Error"
	slug := "error"

	switch code {
	case "invalid_request":
		title, slug = "Invalid request", "invalid-request"
	case "trip_not_found":
		title, slug = "Trip not found", "trip-not-found"
	case "trip_completed":
		title, slug = "Trip completed", "trip-completed"
	case "driver_busy":
		title, slug = "Driver busy", "driver-busy"
	case "internal_error":
		title, slug = "Internal Server Error", "internal-error"
	}

	return api.Problem{
		Type:   "https://tripgo.example/problems/" + slug,
		Title:  title,
		Status: int32(status),
		Detail: &detail,
		Code:   code,
	}
}
