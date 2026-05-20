# Tuya Smart Life integration

This runbook covers Tuya-side setup for `smart-bridge`. Cloud Run deployment and Yandex Smart Home setup are separate steps.

## Prerequisites

- Tuya Smart or Smart Life mobile app account with the target devices already added and working.
- Tuya IoT Platform account. Create it at [iot.tuya.com](https://iot.tuya.com/) if needed.
- Device data center is known.

## Create Cloud Project

1. Open the [Tuya IoT Platform](https://iot.tuya.com/).
2. Go to **Cloud** -> **Development**.
3. Create a Smart Home / Smart Home PaaS cloud project.
4. Select the data center that contains the devices from the Tuya Smart or Smart Life mobile app.
5. Enable the API products required by the project wizard for device discovery, status reads, and device control.

The project data center, OpenAPI endpoint, and linked app account must match.

## Fill smart-bridge Config

Create local config if it does not exist yet:

```sh
cp config.example.yaml config.yaml
```

Copy the project credentials from the Tuya project overview / authorization page:

```text
Access ID / Client ID -> tuya.client_id
Access Secret / Client Secret -> tuya.client_secret
OpenAPI endpoint for the same data center -> tuya.endpoint
```

Example for Europe:

```yaml
tuya:
  endpoint: https://openapi.tuyaeu.com
  client_id: <Access ID / Client ID>
  client_secret: <Access Secret / Client Secret>
```

## Link Mobile App Account

1. In the Tuya cloud project, open **Devices**.
2. Choose **Link App Account** / **Link Devices by App Account**.
3. Select the same mobile app family that owns the devices, usually **Smart Life** or **Tuya Smart**.
4. Scan the QR code from that mobile app account.
5. Confirm that the devices appear in the cloud project device list.

`smart-bridge` can only discover and control devices visible to the linked Tuya cloud project.

## Verify

After saving `config.yaml`, verify through the CLI:

```sh
bin/smart-bridge --config config.yaml devices list
bin/smart-bridge --config config.yaml devices capabilities <device-id>
```

## Troubleshooting

If authentication fails, check credentials and data center.

If authentication works but no devices appear, check the linked app account and whether devices are visible in the Tuya cloud project.

## Sources

- [Manage projects](https://developer.tuya.com/en/docs/iot/manage-projects?_source=02ec1530a4f6dda7f2863fcb220b292a&id=Ka49p0n8vkzm6)
- [Link devices](https://developer.tuya.com/en/docs/iot/link-devices?id=Ka471nu1sfmkl)
