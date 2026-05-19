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
		Handler:           newMux(cfg, newYandexHandler(cfg, gateway, logger)),
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

func newMux(cfg config.Config, yandexHandler http.Handler) http.Handler {
	mux := http.NewServeMux()
	pattern, stripPrefix := yandexMount(cfg.Yandex.PathPrefix)
	mux.Handle(pattern, http.StripPrefix(stripPrefix, yandexHandler))

	return mux
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
