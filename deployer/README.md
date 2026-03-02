# deployer

Manage multiple compose stacks from a single YAML config. Supports status checks, start, stop, and restart for individual or all stacks.

## Install

```bash
cd deployer
go build -o deployer
mv deployer ~/.local/bin/
```

Requires `gopkg.in/yaml.v3`:

```bash
go mod tidy
```

## Usage

```bash
# show status of all stacks
deployer status

# start a specific stack
deployer up vaultwarden

# stop all stacks
deployer down

# restart a stack
deployer restart gitea

# custom config path
deployer -c /etc/deployer.yaml status
```

## Config

Create `deployer.yaml`:

```yaml
runtime: podman
stacks:
  - name: vaultwarden
    path: /opt/stacks/vaultwarden
  - name: gitea
    path: /opt/stacks/gitea
  - name: immich
    path: /opt/stacks/immich
```

The `runtime` field accepts `podman` or `docker` (default: `podman`).

## Commands

| Command | Description |
|---------|-------------|
| `status` | Show running/stopped state of stacks (default) |
| `up` | Start stacks with `compose up -d` |
| `down` | Stop stacks with `compose down` |
| `restart` | Stop then start stacks |

All commands accept an optional stack name to target a single stack.

## Flags

```
-c string       path to config file (default "deployer.yaml")
--no-color      disable colored output
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | success |
| 1 | stack not found |
| 2 | config error or unknown command |

## Example output

```
deployer -> status

  * vaultwarden        running
  * gitea              running
  * immich             stopped
```

## Testing

```bash
go test -v ./...
```

Tests use a mock runner interface to verify logic without calling Docker/Podman.

## License

MIT
