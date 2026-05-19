package cloud

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type routeContextKey struct{}

type loggingRoundTripper struct {
	next   http.RoundTripper
	logger *slog.Logger
}

func (transport *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	enabled := transport.logger.Enabled(req.Context(), slog.LevelDebug)
	var start time.Time
	if enabled {
		start = time.Now()
	}

	resp, err := transport.nextTransport().RoundTrip(req)
	if !enabled {
		return resp, err
	}

	attrs := []any{
		"method", req.Method,
		"route", requestRoute(req),
		"duration", time.Since(start),
	}
	if resp != nil {
		attrs = append(attrs, "status_code", resp.StatusCode)
	}
	if err != nil {
		attrs = append(attrs, "error", err)
	}

	transport.logger.DebugContext(req.Context(), "tuya http request completed", attrs...)

	return resp, err
}

func (transport *loggingRoundTripper) nextTransport() http.RoundTripper {
	if transport.next != nil {
		return transport.next
	}

	return http.DefaultTransport
}

func requestRoute(req *http.Request) string {
	route, _ := req.Context().Value(routeContextKey{}).(string)
	if route == "" {
		return "{unknown}"
	}

	return route
}

func withRoute(ctx context.Context, route string) context.Context {
	return context.WithValue(ctx, routeContextKey{}, route)
}
