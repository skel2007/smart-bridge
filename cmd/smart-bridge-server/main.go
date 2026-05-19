package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/skel2007/smart-bridge/internal/server"
)

const (
	defaultConfigPath = "config.yaml"

	exitOK           = 0
	exitRuntimeError = 1
	exitUsageError   = 2
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	code := run(ctx, os.Args[1:], logger)
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, logger *slog.Logger) int {
	configPath, code, ok := parseFlags(args)
	if !ok {
		return code
	}

	if err := server.Run(ctx, configPath, logger); err != nil {
		logger.ErrorContext(ctx, "http server failed", "error", err)
		return exitRuntimeError
	}

	return exitOK
}

func parseFlags(args []string) (configPath string, code int, ok bool) {
	flags := flag.NewFlagSet("smart-bridge-server", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&configPath, "config", defaultConfigPath, "path to config file")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(os.Stdout, flags)
			return "", exitOK, false
		}

		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printUsage(os.Stderr, flags)
		return "", exitUsageError, false
	}
	if flags.NArg() > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Error: unexpected argument: %s\n\n", flags.Arg(0))
		printUsage(os.Stderr, flags)
		return "", exitUsageError, false
	}

	return configPath, exitOK, true
}

func printUsage(w io.Writer, flags *flag.FlagSet) {
	_, _ = fmt.Fprintln(w, "HTTP server for Yandex Smart Home API.")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Usage: smart-bridge-server [--config config.yaml]")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "  --config string")
	_, _ = fmt.Fprintf(w, "\tpath to config file (default %q)\n", defaultConfigPath)
}
