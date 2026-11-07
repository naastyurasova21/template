package get

import (
	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/model"
)

func toAPITrip(t *model.Trip) api.Trip {
	var finishedAt = t.FinishedAt

	return api.Trip{
		Id:             t.ID,
		UserId:         t.UserID,
		DriverId:       t.DriverID,
		StartPoint:     api.Coordinates{Latitude: t.StartLatitude, Longitude: t.StartLongitude},
		EndPoint:       api.Coordinates{Latitude: t.EndLatitude, Longitude: t.EndLongitude},
		Price:          t.Price,
		Status:         api.TripStatus(t.Status),
		StartedAt:      t.StartedAt,
		FinishedAt:     finishedAt,
		LastPositionAt: t.LastPositionAt,
	}
}
