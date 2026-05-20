# ADR 0004: Logging boundaries

## Status

Accepted.

## Context

The HTTP service is long-running and needs diagnostics for upstream outages and downstream request degradation. The CLI can keep returning errors directly, but shared adapters should not depend on CLI output.

Tuya Cloud retries can produce multiple physical HTTP requests for one logical device operation. Token refresh also puts the refresh token in the URL path, so raw request URL logging can leak credentials.

## Decision

Use `log/slog` for structured logging. Application wiring owns logger construction and passes loggers into adapter boundaries with `WithLogger` options. If no logger is provided, adapters discard logs.

Keep `internal/devices` log-free.

The Tuya Cloud client logs outgoing HTTP attempts at debug level with method, safe route, numeric `duration_ms`, status code, and error when present. It does not log raw refresh-token URLs.

The Yandex handler does not log accepted protocol requests. Cloud Run request logs are the primary access logs. The Yandex handler logs upstream failures that affect a whole request at warn level, partial discovery degradation at warn level, and per-device query/action errors at debug level.

Do not add generic HTTP access logging middleware in the Yandex handler. Generic access logs belong to the platform or HTTP server entrypoint; in Cloud Run, `run.googleapis.com/requests` is the primary access log. Yandex protocol request logs belong at the Yandex boundary only when they add protocol-specific diagnostics such as `X-Request-Id`.

The OAuth compatibility handler logs safe token-client rejection events at warn level. It records only non-secret diagnostic fields such as error code and grant type. It must not log authorization codes, refresh tokens, client secrets, bearer tokens, or raw `Authorization` headers.

The HTTP server entrypoint creates a JSON logger on stderr at debug level by default and passes it into server wiring. Its JSON keys are shaped for Cloud Logging: `severity` instead of `level`, `message` instead of `msg`, and Go `WARN` is emitted as Cloud Logging `WARNING`.

## Consequences

Callers that do not care about logs keep using the existing constructors without options.

The HTTP entrypoint creates one logger and passes it to both Tuya and Yandex wiring without package-level globals.

Tuya transport logging remains close to retries and token lifecycle, while Yandex logs describe user-facing request degradation.
