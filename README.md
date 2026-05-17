# smart-bridge

`smart-bridge` is a Go command-line utility for inspecting and controlling smart home devices.

Current scope is Tuya Cloud device discovery, capability inspection, and basic control from a Smart Life/Tuya project.

## Configuration

Create a local config from the example:

```sh
cp config.example.yaml config.yaml
```

Fill in the Tuya Cloud credentials:

```yaml
tuya:
  endpoint: https://openapi.tuyaeu.com
  client_id: your-client-id
  client_secret: your-client-secret
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

## Development

Run tests:

```sh
go test ./...
```

## License

MIT
