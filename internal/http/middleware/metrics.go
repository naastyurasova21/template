package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func Metrics() func(next http.Handler) http.Handler {
	meter := otel.Meter("trip-service")

	durationHist, _ := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of HTTP server requests"),
	)

	activeRequests, _ := meter.Int64UpDownCounter(
		"http.server.active_requests",
		metric.WithUnit("{request}"),
		metric.WithDescription("Number of active HTTP server requests"),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			activeAttrs := []attribute.KeyValue{
				attribute.String("http.request.method", r.Method),
			}
			activeRequests.Add(r.Context(), 1, metric.WithAttributes(activeAttrs...))

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			activeRequests.Add(r.Context(), -1, metric.WithAttributes(activeAttrs...))

			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unknown"
			}

			statusCode := strconv.Itoa(ww.Status())
			histAttrs := []attribute.KeyValue{
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route),
				attribute.String("http.response.status_code", statusCode),
			}

			durationHist.Record(r.Context(), time.Since(start).Seconds(), metric.WithAttributes(histAttrs...))
		})
	}
}