package create
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/http/domainerror"
	"github.com/naastyurasova21/template/internal/model"
	"github.com/naastyurasova21/template/internal/usecase/trip/create"
)

type Usecase interface {
	Execute(ctx context.Context, in create.Input) (*model.Trip, error)
}

type Handler struct {
	uc     Usecase
	logger *slog.Logger
}

func New(uc Usecase, logger *slog.Logger) *Handler {
	return &Handler{uc: uc, logger: logger}
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, _ api.CreateTripParams) {
	var req api.TripData
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		h.logger.Warn("decode create trip request failed", "err", err)
		writeProblem(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if dec.More() {
		writeProblem(w, http.StatusBadRequest, "unexpected data after JSON body")
		return
	}

	trip, err := h.uc.Execute(r.Context(), create.Input{
		UserID:   req.UserId,
		DriverID: req.DriverId,
		StartLat: req.StartPoint.Latitude,
		StartLon: req.StartPoint.Longitude,
		EndLat:   req.EndPoint.Latitude,
		EndLon:   req.EndPoint.Longitude,
		Price:    req.Price,
	})
	if err != nil {
		status, problem := domainerror.Problem(err)
		if status >= http.StatusInternalServerError {
			h.logger.Error("create trip failed", "err", err)
		}
		writeJSON(w, status, problem)
		return
	}

	w.Header().Set("Location", "/api/v1/trips/"+trip.ID.String())
	writeJSON(w, http.StatusCreated, toAPITrip(trip))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, api.Problem{
		Type:   "https://tripgo.example/problems/invalid-request",
		Title:  "Invalid request",
		Status: int32(status), //nolint:gosec // HTTP status codes are always within int32
		Detail: &detail,
		Code:   "invalid_request",
	})
}
