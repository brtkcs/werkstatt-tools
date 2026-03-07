# hookrelay

Webhook receiver that logs incoming payloads to a JSON lines file. Accepts POST requests with event data and appends structured log entries.

## Install

```bash
cd hookrelay
go build -o hookrelay
mv hookrelay ~/.local/bin/
```

## Usage

```bash
# start with defaults (port 8080, webhooks.log)
hookrelay

# custom port and log file
hookrelay -port 9090 -log /var/log/webhooks.log

# no color (for systemd/logging)
hookrelay --no-color
```

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/webhook` | POST | Receive webhook payload |

## Payload format

```json
{"event":"push","repo":"werkstatt","branch":"main"}
```

The `event` field is required. `repo` and `branch` are optional.

## Testing with curl

```bash
# health check
curl localhost:8080/health

# send a webhook
curl -X POST localhost:8080/webhook \
  -d '{"event":"push","repo":"werkstatt","branch":"main"}'

# check the log
cat webhooks.log
```

## Log format

Each line in the log file is a JSON object:

```json
{"time":"2026-03-07 14:22:05","event":"push","repo":"werkstatt","branch":"main"}
```

## Flags

```
-port string    listen port (default "8080")
-log string     path to log file (default "webhooks.log")
--no-color      disable colored output
```

## Testing

```bash
go test -v ./...
```

Tests use httptest and a mock logger interface.

## License

MIT
