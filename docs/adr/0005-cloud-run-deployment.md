# ADR 0005: Cloud Run deployment

## Status

Accepted.

## Context

smart-bridge is a **Personal Bridge** that exposes the Yandex Smart Home REST API and calls Tuya Cloud over outbound HTTPS. It needs a public HTTPS endpoint, but it does not currently need Kubernetes-specific APIs or cluster operations.

## Decision

Deploy the first public `smart-bridge-server` as a container on Google Cloud Run in `europe-west1`, using the default Cloud Run `*.run.app` HTTPS endpoint. Do not introduce Kubernetes until smart-bridge needs cluster-native operations or grows into multiple services.

Allow public Cloud Run ingress and unauthenticated platform access. Yandex protocol endpoints use the application-level bearer token; Cloud Run IAM authentication is not compatible with Yandex calling the bridge as a regular HTTPS webhook client.

Build the server image with Cloud Build from the repository-owned Dockerfile, push it to Artifact Registry, and deploy Cloud Run from that image. Tag images by source revision, such as a short Git SHA; do not rely on `latest` as the deployed reference.

Mount one Secret Manager secret containing the full runtime `config.yaml` at `/secrets/smart-bridge/config.yaml`. The Cloud Run config uses `http.port: 8080`; local development may use `http.port: 0`.

Run Cloud Run with `min-instances=0`, `max-instances=1`, default concurrency, and request timeout `60s`. The single-instance cap is a first-deployment cost and rate-limit guard, not a correctness requirement; multiple instances are acceptable while smart-bridge remains stateless.

Use a dedicated Cloud Run runtime service account with Secret Manager access only to the `smart-bridge-config` secret. Rotate config or credentials by adding a new secret version and deploying a new Cloud Run revision pinned to that numeric version.

Use Cloud Logging from JSON stdout/stderr logs at debug level. Do not add alerts or log sinks until real traffic shows an operational need.

Expose unauthenticated `GET /health` outside `yandex.path_prefix`. It checks only that the HTTP process is up and does not call Tuya Cloud. Avoid `z`-suffixed health paths because Cloud Run reserves some URL paths ending with `z`. Post-deployment smoke still uses the Yandex protocol endpoints with `Authorization` and `X-Request-Id`.

## Consequences

Cloud Run is easier to operate than Kubernetes for this single stateless service, but it makes the first deployment GCP-specific.

The default `*.run.app` endpoint avoids custom domain setup, but changing cloud project, region, or platform later will require updating the Yandex callback URL.

The mounted config secret avoids code changes for environment-based configuration, but non-secret config and secret values live in the same Secret Manager entry for now.
