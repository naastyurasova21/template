package container

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naastyurasova21/template/api"
	"github.com/naastyurasova21/template/internal/config"
	"github.com/naastyurasova21/template/internal/db"
	"github.com/naastyurasova21/template/internal/http/domainerror"
	tripcreate "github.com/naastyurasova21/template/internal/http/trip/create"
	tripfinish "github.com/naastyurasova21/template/internal/http/trip/finish"
	tripget "github.com/naastyurasova21/template/internal/http/trip/get"
	"github.com/naastyurasova21/template/internal/http/service"
	"github.com/naastyurasova21/template/internal/infra/postgres/transaction"
	"github.com/naastyurasova21/template/internal/repository/trip"
	usecasecreate "github.com/naastyurasova21/template/internal/usecase/trip/create"
	usecasefinish "github.com/naastyurasova21/template/internal/usecase/trip/finish"
	usecaseget "github.com/naastyurasova21/template/internal/usecase/trip/get"
)

type Container struct {
	Server *http.Server
	Pool   *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Container, error) {
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init db pool: %w", err)
	}

	txManager := transaction.New(pool)

	tripRepo := trip.New(pool, cfg.DatabaseQueryTimeout)

	createUC := usecasecreate.New(tripRepo, txManager)
	getUC := usecaseget.New(tripRepo)
	finishUC := usecasefinish.New(tripRepo, txManager)

	createHandler := tripcreate.New(createUC, logger)
	getHandler := tripget.New(getUC, logger)
	finishHandler := tripfinish.New(finishUC, logger)
	serviceHandler := service.New(pool, cfg.DatabaseQueryTimeout)

	serverImpl := &serverImpl{
		create:  createHandler,
		get:     getHandler,
		finish:  finishHandler,
		service: serviceHandler,
	}

	router := chi.NewRouter()
	apiHandler := api.HandlerWithOptions(serverImpl, api.ChiServerOptions{
		BaseRouter:       router,
		ErrorHandlerFunc: domainerror.ErrorHandler,
	})

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           apiHandler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Container{
		Server: httpServer,
		Pool:   pool,
	}, nil
}
