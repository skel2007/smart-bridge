# ADR 0004: Logging boundaries

## Status

Accepted.

## Context

The HTTP service is long-running and needs diagnostics for upstream outages and downstream request degradation. The CLI can keep returning errors directly, but shared adapters should not depend on CLI output.

Tuya Cloud retries can produce multiple physical HTTP requests for one logical device operation. Token refresh also puts the refresh token in the URL path, so raw request URL logging can leak credentials.

## Decision

Use `log/slog` for structured logging. Application wiring owns logger construction and passes loggers into adapter boundaries with `WithLogger` options. If no logger is provided, adapters discard logs.

Keep `internal/devices` log-free.

The Tuya Cloud client logs outgoing HTTP attempts at debug level with method, safe route, duration, status code, and error when present. It does not log raw refresh-token URLs.

The Yandex handler logs accepted protocol requests at info level with the Yandex request ID, method, and handler-local path. It logs upstream failures that affect a whole request at warn level, partial discovery degradation at warn level, and per-device query/action errors at debug level.

Do not add generic HTTP access logging middleware in the Yandex handler. Generic access logs belong with the HTTP server entrypoint; Yandex protocol request logs belong at the Yandex boundary because `X-Request-Id` is part of that protocol.

The HTTP server entrypoint creates a JSON logger on stderr at debug level by default and passes it into server wiring.

## Consequences

Callers that do not care about logs keep using the existing constructors without options.

The HTTP entrypoint creates one logger and passes it to both Tuya and Yandex wiring without package-level globals.

Tuya transport logging remains close to retries and token lifecycle, while Yandex logs describe user-facing request degradation.
