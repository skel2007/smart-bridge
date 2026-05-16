# smart-bridge

`smart-bridge` is a Go command-line utility for working with smart home device integrations.

Current scope:

- Cobra-based CLI entrypoint: `cmd/smart-bridge`
- device command skeleton: `smart-bridge devices list`
- Tuya Cloud support will be added as the first integration

## Development

Run tests:

```sh
go test ./...
```

Build the CLI binary:

```sh
go build -o bin/smart-bridge ./cmd/smart-bridge
```

## License

MIT
