# ADR 0006: Single-user OAuth compatibility

## Status

Accepted.

## Context

Yandex Smart Home account linking expects OAuth 2.0 authorization code flow. ADR 0003 originally kept OAuth endpoints out of smart-bridge, but the deployed private Yandex skill cannot be linked without filling authorization and token endpoint settings.

smart-bridge remains a **Personal Bridge**, not a public multi-user provider. It has one configured bridge-side user and one deployed Yandex API surface.

## Decision

Implement a single-user OAuth compatibility layer inside `internal/yandex`. It is not a public OAuth provider.

Expose OAuth endpoints under `yandex.path_prefix`:

- `GET /oauth/authorize`
- `POST /oauth/token`

The authorization endpoint auto-approves the configured user. It accepts `response_type=code`, `client_id`, `redirect_uri`, optional `state`, and optional `scope`; it requires an absolute HTTPS `redirect_uri` and redirects back with a signed short-lived code.

The token endpoint accepts client authentication only through HTTP Basic. It supports `grant_type=authorization_code` and `grant_type=refresh_token`. For authorization code exchange, it requires `redirect_uri` and checks exact match with the signed code.

The access token returned to Yandex is the configured `yandex.bearer_token`. Authorization codes and refresh tokens are stateless signed tokens using `yandex.oauth.client_secret`; authorization codes expire after 5 minutes, and refresh tokens use a rolling 365-day lifetime. `scope` is optional and is returned only when requested.

## Consequences

The account link survives Cloud Run restarts and scale-to-zero without Firestore or another token store.

Authorization codes are signed and short-lived, but they are not one-time-use. This is an intentional compromise for a private **Personal Bridge**.

Revocation means rotating `yandex.oauth.client_secret` or `yandex.bearer_token`, uploading a new config secret version, and deploying a new Cloud Run revision.

This compatibility layer may be insufficient for public catalog moderation. If smart-bridge becomes a public multi-user provider, replace it with a real OAuth implementation backed by persistent storage and user consent.
