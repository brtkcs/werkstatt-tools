# stackctl

Lightweight REST API that exposes system information and container status. Supports both Docker and Podman.

## Install

    cd stackctl
    go build -o stackctl

## Usage

    stackctl                          # defaults: port 8080, podman
    stackctl -port 9090 -runtime docker
    stackctl --no-color               # for logging/pipes

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| /health | GET | Health check with runtime info |
| /stacks | GET | Running containers as JSON |
| /info | GET | System info (uname, uptime, memory, disk) |

## Flags

    -port string      listen port (default "8080")
    -runtime string   container runtime: docker or podman (default "podman")
    --no-color        disable colored output

## Testing

    go test -v ./...

## License

MIT
