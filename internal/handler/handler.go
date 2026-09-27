package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/repository"
	"github.com/naastyurasova21/template/internal/txmanager"
)

type Handler struct {
	tripRepo  *repository.TripRepository
	txManager txmanager.TxManager
	pool      *pgxpool.Pool
	logger    *slog.Logger
}

func New(
	tripRepo *repository.TripRepository,
	txManager txmanager.TxManager,
	pool *pgxpool.Pool,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		tripRepo:  tripRepo,
		txManager: txManager,
		pool:      pool,
		logger:    logger,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.pool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, api.HealthResponse{Status: api.Unavailable})
		return
	}
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: api.Ok})
}

func (h *Handler) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	// params.IdempotencyKey — для задания со звёздочкой

	var data api.TripData
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // additionalProperties: false
	if err := dec.Decode(&data); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON body: "+err.Error())
		return
	}

	if err := validateTripData(data); err != nil {
		writeProblem(w, r, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	var trip *api.Trip
	err := h.txManager.Do(r.Context(), func(ctx context.Context) error {
		var err error
		trip, err = h.tripRepo.Create(ctx, data)
		return err
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDriverBusy):
			writeProblem(w, r, http.StatusConflict, "driver_busy", "Driver already has an active trip")
		default:
			h.logger.Error("create trip failed", "err", err)
			writeProblem(w, r, http.StatusInternalServerError, "internal_error", "Internal error")
		}
		return
	}

	w.Header().Set("Location", "/api/v1/trips/"+trip.Id.String())
	writeJSON(w, http.StatusCreated, trip)
}

func (h *Handler) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.tripRepo.GetByID(r.Context(), tripId)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTripNotFound):
			writeProblem(w, r, http.StatusNotFound, "trip_not_found", "Trip not found")
		default:
			h.logger.Error("get trip failed", "err", err)
			writeProblem(w, r, http.StatusInternalServerError, "internal_error", "Internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, trip)
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	var trip *api.Trip
	err := h.txManager.Do(r.Context(), func(ctx context.Context) error {
		var err error
		trip, err = h.tripRepo.Finish(ctx, tripId)
		return err
	})
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTripNotFound):
			writeProblem(w, r, http.StatusNotFound, "trip_not_found", "Trip not found")
		case errors.Is(err, repository.ErrTripAlreadyCompleted):
			writeProblem(w, r, http.StatusConflict, "trip_completed", "Trip already completed")
		default:
			h.logger.Error("finish trip failed", "err", err)
			writeProblem(w, r, http.StatusInternalServerError, "internal_error", "Internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, trip)
}

func (h *Handler) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	writeProblem(w, r, http.StatusNotImplemented, "not_implemented", "Trip positions are not implemented yet")
}

func (h *Handler) ListTripPositions(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	writeProblem(w, r, http.StatusNotImplemented, "not_implemented", "Trip positions are not implemented yet")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func validateTripData(data api.TripData) error {
	if data.UserId == uuid.Nil {
		return errors.New("user_id is required")
	}
	if data.DriverId == uuid.Nil {
		return errors.New("driver_id is required")
	}
	if err := validateCoordinates("start_point", data.StartPoint); err != nil {
		return err
	}
	if err := validateCoordinates("end_point", data.EndPoint); err != nil {
		return err
	}
	if data.Price < 0 {
		return errors.New("price must be >= 0")
	}
	return nil
}

func validateCoordinates(field string, c api.Coordinates) error {
	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf("%s.latitude=%v out of range [-90, 90]", field, c.Latitude)
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf("%s.longitude=%v out of range [-180, 180]", field, c.Longitude)
	}
	return nil
}
