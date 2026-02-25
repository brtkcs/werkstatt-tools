# werkstatt

A collection of Go-based DevOps tools built for real-world homelab and infrastructure management.

Each tool solves a specific problem I encountered while managing self-hosted services across multiple machines. Built with Go, designed for single-binary deployment.

## Tools

### CLI Tools

| Tool | Description |
|------|-------------|
| **envcheck** | `.env` file validator – compares against `.env.example`, catches missing keys, duplicates, and empty values |
| **portspy** | Concurrent TCP port scanner – scans multiple ports simultaneously using goroutines |
| **sshping** | SSH host availability checker – reads targets from YAML config, checks connectivity |

### Web / API

| Tool | Description |
|------|-------------|
| **hookrelay** | Webhook receiver and logger – accepts POST webhooks, logs to JSON Lines file |
| **stackctl** | System info REST API – exposes host metrics (uname, memory, disk, uptime) over HTTP |
| **kvault** | Key-value store with REST API – in-memory with JSON file persistence |

### System Tools

| Tool | Description |
|------|-------------|
| **dmon** | Process monitor – watches named processes, reports state changes, YAML configurable |
| **netmapper** | Network discovery tool – scans subnets, identifies live hosts and open ports |
| **deployer** | Compose stack manager – start/stop/status for Docker and Podman stacks from a single config |

## Quick start

Each tool is a standalone Go project. To build any of them:

```bash
cd envcheck
go build -o envcheck
./envcheck
```

Or run directly:

```bash
go run main.go
```

## Design principles

- **Single binary** – `go build` produces one executable, no dependencies
- **YAML/JSON config** – no hardcoded values, easy to adapt
- **Minimal dependencies** – standard library where possible
- **Self-hosted first** – built for homelab infrastructure, not cloud

## Tech stack

- **Go 1.22+**
- Standard library: `net/http`, `os/exec`, `encoding/json`, `flag`, `sync`
- External: `gopkg.in/yaml.v3` (YAML parsing)

## Project structure

```
werkstatt/
├── envcheck/       # .env validator
├── portspy/        # concurrent port scanner
├── sshping/        # SSH host checker
├── hookrelay/      # webhook receiver + logger
├── stackctl/       # system info REST API
├── kvault/         # key-value store API
├── dmon/           # process monitor
├── netmapper/      # network discovery
├── deployer/       # compose stack manager
└── README.md
```

## Related

- [dotfiles](https://github.com/brtkcs/dotfiles) – Neovim, bash, starship config
- [homelab-monitoring](https://github.com/brtkcs/homelab-monitoring) – Prometheus + Grafana stack

## License

MIT
