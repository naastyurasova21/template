package finish
//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/http/domainerror"
	"github.com/naastyurasova21/template/internal/model"
)

type Usecase interface {
	Execute(ctx context.Context, id uuid.UUID) (*model.Trip, error)
}

type Handler struct {
	uc     Usecase
	logger *slog.Logger
}

func New(uc Usecase, logger *slog.Logger) *Handler {
	return &Handler{uc: uc, logger: logger}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.uc.Execute(r.Context(), tripId)
	if err != nil {
		status, problem := domainerror.Problem(err)
		if status >= http.StatusInternalServerError {
			h.logger.Error("finish trip failed", "err", err)
		}
		writeJSON(w, status, problem)
		return
	}

	writeJSON(w, http.StatusOK, toAPITrip(trip))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
