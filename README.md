# smart-bridge

`smart-bridge` is a Go utility for inspecting, controlling, and exposing smart home devices.

Current scope is Tuya Cloud device discovery, capability inspection, basic control from a Smart Life/Tuya project, and a Yandex Smart Home HTTP server.

## Configuration

Create a local config from the example:

```sh
cp config.example.yaml config.yaml
```

Fill in the runtime and platform settings:

```yaml
http:
  port: 0

tuya:
  endpoint: https://openapi.tuyaeu.com
  client_id: your-client-id
  client_secret: your-client-secret

yandex:
  user_id: your-stable-user-id
  bearer_token: your-strong-bearer-token
  path_prefix: /api/yandex
```

`config.yaml` contains local secrets and is ignored by git.

## Usage

Build the CLI binary:

```sh
go build -o bin/smart-bridge ./cmd/smart-bridge
```

Inspect devices and capabilities:

```sh
bin/smart-bridge devices list
bin/smart-bridge devices list --json
bin/smart-bridge devices capabilities <device-id>
bin/smart-bridge devices capabilities <device-id> --json
bin/smart-bridge --config path/to/config.yaml devices list
```

Set device capabilities:

```sh
bin/smart-bridge devices set power <device-id> on
bin/smart-bridge devices set brightness <device-id> 50
bin/smart-bridge devices set color-temperature <device-id> 75
bin/smart-bridge devices set color <device-id> --hue 120 --saturation 80 --value 90
bin/smart-bridge devices set mode <device-id> white
```

Use `--json` with `devices list` or `devices capabilities` to print JSON.

Known range capabilities use vendor-neutral `0..100` values in output. For example, Tuya brightness and color temperature level ranges are normalized before printing or JSON encoding.

`color` is available only when the device exposes a color capability. Some Tuya lights expose only power, brightness, color temperature, and mode/scene controls.

Build the Yandex Smart Home HTTP server binary:

```sh
go build -o bin/smart-bridge-server ./cmd/smart-bridge-server
```

Run the HTTP server:

```sh
bin/smart-bridge-server --config config.yaml
```

The server listens on `http.port`. The default example uses `0`, so the OS chooses an available local port and the actual address is written to JSON logs on stderr.

Yandex requests are mounted under `yandex.path_prefix`. With the default config, the local discovery path is `/api/yandex/v1.0/user/devices`.

Before exposing the server outside local development, deploy it behind HTTPS and configure a strong bearer token. Domain name, TLS, DNS, and reverse proxy or load balancer routing are outside smart-bridge.

## Development

Run tests:

```sh
go test ./...
```

## License

MIT
