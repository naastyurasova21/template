package container

import (
	"net/http"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/http/service"
	tripcreate "github.com/naastyurasova21/template/internal/http/trip/create"
	tripfinish "github.com/naastyurasova21/template/internal/http/trip/finish"
	tripget "github.com/naastyurasova21/template/internal/http/trip/get"
)

type serverImpl struct {
	create  *tripcreate.Handler
	get     *tripget.Handler
	finish  *tripfinish.Handler
	service *service.Handler
}

func (s *serverImpl) CreateTrip(w http.ResponseWriter, r *http.Request, params api.CreateTripParams) {
	s.create.CreateTrip(w, r, params)
}

func (s *serverImpl) GetTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	s.get.GetTrip(w, r, tripId)
}

func (s *serverImpl) FinishTrip(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	s.finish.FinishTrip(w, r, tripId)
}

func (s *serverImpl) Health(w http.ResponseWriter, r *http.Request) {
	s.service.Health(w, r)
}

func (s *serverImpl) Ready(w http.ResponseWriter, r *http.Request) {
	s.service.Ready(w, r)
}

func (s *serverImpl) CreateTripPosition(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *serverImpl) ListTripPositions(w http.ResponseWriter, r *http.Request, tripId api.TripId) {
	w.WriteHeader(http.StatusNotImplemented)
}
