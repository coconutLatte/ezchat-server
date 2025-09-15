# ezchat-server

## Run

Build and run:

```bash
go run ./...
```

## CLI

The server uses Cobra CLI and Viper for configuration.

- Global flag `--config` (or `-c`) specifies config file path.
- If omitted, Viper searches default locations: `./configs`, current directory, and reads environment variables with prefix `EZCHAT`.

Example:

```bash
ezchat-server --config ./config/config.yaml
```

### Config example

See `config/config.example.yaml` for structure.