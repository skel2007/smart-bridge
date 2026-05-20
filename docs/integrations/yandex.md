# Yandex Smart Home integration

This runbook covers Yandex-side setup for a private Smart Home skill. Tuya setup and Cloud Run deployment are separate steps.

## Prerequisites

- Tuya integration works, and `smart-bridge` can list the target devices.
- Cloud Run service is deployed.
- Current Cloud Run service URL is known.
- `config.yaml` has Yandex protocol and OAuth settings filled.

Use these placeholders below:

```text
YANDEX_BASE_URL = <Cloud Run service URL> + <yandex.path_prefix>
```

For example, if `yandex.path_prefix` is `/api/yandex`, the backend URL is:

```text
https://<service>.run.app/api/yandex
```

Do not include `/v1.0`, `/user/devices`, `/query`, or `/action` in `YANDEX_BASE_URL`. Yandex appends Smart Home protocol paths itself.

## Create Skill Draft

1. Open the [Yandex Dialogs developer console](https://dialogs.yandex.ru/developer/).
2. Create a new dialog of type **Smart home**.
3. Open the draft settings.

## Fill General Settings

Fill the required name, description, icon, and display fields with private/test wording. This is a personal bridge, not a public provider.

## Fill Access Settings

Set access type to **Private**.

Moderator notes can be left empty for a private skill.

## Fill Backend Settings

Set the Smart Home backend URL / endpoint URL to:

```text
${YANDEX_BASE_URL}
```

With the default `yandex.path_prefix`, Yandex will call discovery at:

```text
${YANDEX_BASE_URL}/v1.0/user/devices
```

## Fill Account Linking

smart-bridge provides a single-user OAuth compatibility layer for Yandex account linking. Fill the account-linking page like this:

```text
Идентификатор приложения / Client Identifier:
  value of yandex.oauth.client_id

Секрет приложения / Client Password:
  value of yandex.oauth.client_secret

URL авторизации / API authorization endpoint:
  ${YANDEX_BASE_URL}/oauth/authorize

URL для получения токена / Token Endpoint:
  ${YANDEX_BASE_URL}/oauth/token

URL для обновления токена / Refreshing an Access Token:
  ${YANDEX_BASE_URL}/oauth/token

Идентификатор группы действий / Access Token Scope:
  leave empty
```

The scope field is optional. If Yandex sends a scope value, smart-bridge echoes it in OAuth responses but does not use it for authorization.

The OAuth compatibility layer auto-approves the configured client, returns `yandex.bearer_token` as `access_token`, and Yandex then uses that bearer token for Smart Home protocol requests.

## Publish Skill

1. Save the draft.
2. Run Yandex console validation if available.
3. Publish the private skill.
4. Republish the skill if `YANDEX_BASE_URL` or OAuth endpoint settings change. Relink the skill in the Yandex app if OAuth credentials or `yandex.bearer_token` change.

## Verify

1. Open the Yandex Smart Home mobile app with the same Yandex account that owns the private skill.
2. Add a new smart home provider/device.
3. Find the private skill by name.
4. Complete account linking.
5. Update the device list and confirm that the expected devices appear.
6. Test discovery/query first; test on/off actions only when intentional.

## Troubleshooting

If account linking fails before the token request, check the authorization endpoint URL and that it uses the current `YANDEX_BASE_URL`.

If the token request fails, check `yandex.oauth.client_id`, `yandex.oauth.client_secret`, and Cloud Logging entries for `/oauth/token`.

If linking succeeds but Yandex device requests return `401`, check that the deployed config uses the same `yandex.bearer_token` returned by the OAuth compatibility layer.

If no devices appear, check the Tuya integration first with `smart-bridge devices list`, then check the deployed `${YANDEX_BASE_URL}/v1.0/user/devices` response.

## Sources

- [Yandex Dialogs developer console](https://dialogs.yandex.ru/developer/)
- [Creation and settings](https://yandex.ru/dev/dialogs/smart-home/doc/en/start)
- [Operating protocol](https://yandex.ru/dev/dialogs/smart-home/doc/en/reference/resources)
- [Authorization in a skill](https://yandex.ru/dev/dialogs/smart-home/doc/en/auth/how-it-works)
- [Private skill access](https://yandex.ru/dev/dialogs/smart-home/doc/en/access)
- [Skill testing](https://yandex.ru/dev/dialogs/smart-home/doc/en/testing)
