package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/skel2007/smart-bridge/internal/config"
	"github.com/skel2007/smart-bridge/internal/devices"
	"github.com/skel2007/smart-bridge/internal/tuya"
	"github.com/skel2007/smart-bridge/internal/yandex"
	yandexoauth "github.com/skel2007/smart-bridge/internal/yandex/oauth"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 5 * time.Second
)

func Run(ctx context.Context, configPath string, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	gateway := newTuyaGateway(cfg, logger)
	httpServer := &http.Server{
		Handler: newMux(cfg, yandexHandlers{
			oauth:    newYandexOAuthHandler(cfg, logger),
			protocol: newYandexHandler(cfg, gateway, logger),
		}),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTP.Port))
	if err != nil {
		return fmt.Errorf("listen http server: %w", err)
	}
	logger.InfoContext(ctx, "http server listening", "addr", listener.Addr().String())

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("serve http: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		return nil
	}
}

type yandexHandlers struct {
	oauth    http.Handler
	protocol http.Handler
}

func newMux(cfg config.Config, handlers yandexHandlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", serveHealth)

	pattern, stripPrefix := yandexMount(cfg.Yandex.PathPrefix)
	mux.Handle(pattern, http.StripPrefix(stripPrefix, newYandexMux(handlers)))

	return mux
}

func newYandexMux(handlers yandexHandlers) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/oauth/", http.StripPrefix("/oauth", handlers.oauth))
	mux.Handle("/", handlers.protocol)

	return mux
}

func serveHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func yandexMount(pathPrefix string) (pattern string, stripPrefix string) {
	if pathPrefix == "/" {
		return "/", ""
	}

	return pathPrefix + "/", pathPrefix
}

func newTuyaGateway(cfg config.Config, logger *slog.Logger) devices.DeviceGateway {
	return tuya.NewGateway(
		tuya.Credentials{
			Endpoint:     cfg.Tuya.Endpoint,
			ClientID:     cfg.Tuya.ClientID,
			ClientSecret: cfg.Tuya.ClientSecret,
		},
		tuya.WithLogger(logger),
		tuya.WithSpecificationCache(),
	)
}

func newYandexHandler(cfg config.Config, gateway devices.DeviceGateway, logger *slog.Logger) http.Handler {
	return yandex.NewHandler(
		gateway,
		yandex.HandlerConfig{
			UserID:      cfg.Yandex.UserID,
			BearerToken: cfg.Yandex.BearerToken,
		},
		yandex.WithLogger(logger),
	)
}

func newYandexOAuthHandler(cfg config.Config, logger *slog.Logger) http.Handler {
	return yandexoauth.NewHandler(
		yandexoauth.Config{
			ClientID:          cfg.Yandex.OAuth.ClientID,
			ClientSecret:      cfg.Yandex.OAuth.ClientSecret,
			StaticAccessToken: cfg.Yandex.BearerToken,
		},
		yandexoauth.WithLogger(logger),
	)
}
