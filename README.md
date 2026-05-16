# smart-bridge

`smart-bridge` is a Go command-line utility for inspecting smart home devices.

Current scope is Tuya Cloud device discovery from a Smart Life/Tuya project.

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

List devices:

```sh
bin/smart-bridge devices list
```

Print devices as JSON:

```sh
bin/smart-bridge devices list --json
```

List capabilities for a device:

```sh
bin/smart-bridge devices capabilities <device-id>
```

Print capabilities as JSON:

```sh
bin/smart-bridge devices capabilities <device-id> --json
```

Use a custom config path:

```sh
bin/smart-bridge --config path/to/config.yaml devices list
```

Known range capabilities use vendor-neutral `0..100` values in output. For example, Tuya brightness and color temperature level ranges are normalized before printing or JSON encoding.

## Development

Run tests:

```sh
go test ./...
```

## License

MIT
