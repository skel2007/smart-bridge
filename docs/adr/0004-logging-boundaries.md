# ADR 0004: Logging boundaries

## Status

Accepted.

## Context

The future HTTP service will be long-running and needs diagnostics for upstream outages and downstream request degradation. The CLI can keep returning errors directly, but shared adapters should not depend on CLI output.

Tuya Cloud retries can produce multiple physical HTTP requests for one logical device operation. Token refresh also puts the refresh token in the URL path, so raw request URL logging can leak credentials.

## Decision

Use `log/slog` for structured logging. Application wiring owns logger construction and passes loggers into adapter boundaries with `WithLogger` options. If no logger is provided, adapters discard logs.

Keep `internal/devices` log-free.

The Tuya Cloud client logs outgoing HTTP attempts at debug level with method, safe route, duration, status code, and error when present. It does not log raw refresh-token URLs.

The Yandex handler logs upstream failures that affect a whole request at warn level, partial discovery degradation at warn level, and per-device query/action errors at debug level.

Do not add access logging middleware in the Yandex handler. Request access logs belong with the future HTTP server entrypoint.

## Consequences

Callers that do not care about logs keep using the existing constructors without options.

The HTTP entrypoint can later create one logger and pass it to both Tuya and Yandex wiring without package-level globals.

Tuya transport logging remains close to retries and token lifecycle, while Yandex logs describe user-facing request degradation.
