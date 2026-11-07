package finish

import (
	"context"
	"encoding/json"
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
	uc Usecase
}

func New(uc Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	trip, err := h.uc.Execute(r.Context(), tripId)
	if err != nil {
		status, problem := domainerror.Problem(err)
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
