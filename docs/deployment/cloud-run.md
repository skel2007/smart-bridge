# Cloud Run deployment

This runbook deploys `smart-bridge-server` to Cloud Run from an Artifact Registry image built by Cloud Build. It follows [ADR 0005](../adr/0005-cloud-run-deployment.md).

Tuya and Yandex console setup is out of scope. This runbook assumes valid integration credentials in local `config.yaml`.

## Constants

```sh
export PROJECT_ID=your-gcp-project-id
export REGION=europe-west1
export SERVICE=smart-bridge
export AR_REPO=smart-bridge
export IMAGE_NAME=smart-bridge-server
export SECRET_NAME=smart-bridge-config
export RUNTIME_SA=smart-bridge-runner
export TAG=$(git rev-parse --short HEAD)
export IMAGE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${AR_REPO}/${IMAGE_NAME}:${TAG}"
```

## One-time GCP setup

```sh
gcloud services enable \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  run.googleapis.com \
  secretmanager.googleapis.com \
  --project "${PROJECT_ID}"
```

```sh
gcloud artifacts repositories create "${AR_REPO}" \
  --repository-format docker \
  --location "${REGION}" \
  --project "${PROJECT_ID}" \
  --description "smart-bridge container images"
```

The first manual build uses the default Cloud Build identity. If it cannot push to Artifact Registry, grant that build identity `roles/artifactregistry.writer`.

Create the runtime service account:

```sh
gcloud iam service-accounts create "${RUNTIME_SA}" \
  --project "${PROJECT_ID}" \
  --display-name "smart-bridge Cloud Run runtime"
```

Create the config secret:

```sh
gcloud secrets create "${SECRET_NAME}" \
  --project "${PROJECT_ID}" \
  --replication-policy automatic
```

Grant the runtime service account access to that secret:

```sh
gcloud secrets add-iam-policy-binding "${SECRET_NAME}" \
  --project "${PROJECT_ID}" \
  --member "serviceAccount:${RUNTIME_SA}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/secretmanager.secretAccessor
```

## Config secret

Create local `config.yaml` from `config.example.yaml`. Keep `http.port: 8080`, then add it as a pinned secret version:

```sh
export SECRET_VERSION=$(gcloud secrets versions add "${SECRET_NAME}" \
  --project "${PROJECT_ID}" \
  --data-file config.yaml \
  --format "value(name.basename())")
export REVISION_SUFFIX="${TAG}-${SECRET_VERSION}"
```

smart-bridge reads config only at startup. After config changes, add a new secret version and deploy a new revision pinned to it.

## Optional local container preflight

If Docker is available, verify the Cloud Run config path locally:

```sh
docker build -t smart-bridge-server:local .
docker run --rm \
  -p 8080:8080 \
  -v "$PWD/config.yaml:/secrets/smart-bridge/config.yaml:ro" \
  smart-bridge-server:local
```

In another shell:

```sh
curl -i http://127.0.0.1:8080/health
```

## Build image

```sh
gcloud builds submit \
  --project "${PROJECT_ID}" \
  --tag "${IMAGE}" \
  .
```

## Deploy Cloud Run

```sh
gcloud run deploy "${SERVICE}" \
  --project "${PROJECT_ID}" \
  --region "${REGION}" \
  --image "${IMAGE}" \
  --revision-suffix "${REVISION_SUFFIX}" \
  --port 8080 \
  --service-account "${RUNTIME_SA}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --allow-unauthenticated \
  --ingress all \
  --min-instances 0 \
  --max-instances 1 \
  --timeout 60s \
  --update-secrets /secrets/smart-bridge/config.yaml="${SECRET_NAME}:${SECRET_VERSION}"
```

The deploy command returns the service URL. With the default config, the discovery and OAuth paths are:

```text
https://SERVICE_URL/api/yandex/v1.0/user/devices
https://SERVICE_URL/api/yandex/oauth/authorize
https://SERVICE_URL/api/yandex/oauth/token
```

To read the current service URL later:

```sh
gcloud run services describe "${SERVICE}" \
  --project "${PROJECT_ID}" \
  --region "${REGION}" \
  --format "value(status.url)"
```

## Smoke checks

```sh
export SERVICE_URL=https://your-cloud-run-url
export YANDEX_BEARER_TOKEN=your-strong-bearer-token
export YANDEX_OAUTH_CLIENT_ID=smart-bridge
export YANDEX_OAUTH_CLIENT_SECRET=your-oauth-client-secret
```

```sh
curl -i "${SERVICE_URL}/health"
```

```sh
curl -i \
  -H "X-Request-Id: smoke-wrong-bearer" \
  -H "Authorization: Bearer wrong-token" \
  "${SERVICE_URL}/api/yandex/v1.0/user/devices"
```

```sh
curl -i \
  -H "X-Request-Id: smoke-devices" \
  -H "Authorization: Bearer ${YANDEX_BEARER_TOKEN}" \
  "${SERVICE_URL}/api/yandex/v1.0/user/devices"
```

```sh
curl -i \
  -H "X-Request-Id: smoke-query" \
  -H "Authorization: Bearer ${YANDEX_BEARER_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"devices":[{"id":"your-device-id"}]}' \
  "${SERVICE_URL}/api/yandex/v1.0/user/devices/query"
```

```sh
curl -i -G \
  --data-urlencode "response_type=code" \
  --data-urlencode "client_id=${YANDEX_OAUTH_CLIENT_ID}" \
  --data-urlencode "redirect_uri=https://example.com/callback" \
  --data-urlencode "state=smoke-state" \
  "${SERVICE_URL}/api/yandex/oauth/authorize"
```

```sh
curl -i \
  -u "${YANDEX_OAUTH_CLIENT_ID}:${YANDEX_OAUTH_CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data "grant_type=refresh_token&refresh_token=wrong-token" \
  "${SERVICE_URL}/api/yandex/oauth/token"
```

Inspect Cloud Logging for `http server listening` and any Tuya or Yandex degradation warnings.

## Rollback

```sh
gcloud run revisions list \
  --project "${PROJECT_ID}" \
  --region "${REGION}" \
  --service "${SERVICE}"
```

```sh
gcloud run services update-traffic "${SERVICE}" \
  --project "${PROJECT_ID}" \
  --region "${REGION}" \
  --to-revisions REVISION_NAME=100
```

Run the smoke checks again after rollback.
