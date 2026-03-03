# sshping

Check host reachability from a YAML config. Pings all hosts concurrently and reports status with response times. Returns exit code 1 if any host is unreachable.

## Install

```bash
cd sshping
go build -o sshping
mv sshping ~/.local/bin/
```

Requires `gopkg.in/yaml.v3`:

```bash
go mod tidy
```

## Usage

```bash
# check hosts from default config
sshping

# custom config and timeout
sshping -c /etc/sshping/hosts.yaml -timeout 5

# no color (for cron/logging)
sshping --no-color
```

## Config

Create `hosts.yaml`:

```yaml
hosts:
  - name: media server
    address: 10.0.0.19
    port: 22
  - name: router
    address: 10.0.0.1
    port: 22
  - name: localhost
    address: 127.0.0.1
    port: 631
```

## Example output

```
sshping -> 3 hosts

  * media server         10.0.0.19:22    12ms
  * router               10.0.0.1:22     3ms
  ✗ localhost             127.0.0.1:631   unreachable

  2/3 reachable
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | all hosts reachable |
| 1 | one or more hosts unreachable |
| 2 | config error |

## Flags

```
-c string       path to hosts config (default "hosts.yaml")
-timeout int    timeout in seconds (default 2)
--no-color      disable colored output
```

## Testing

```bash
go test -v ./...
```

Tests use a mock dialer interface to verify logic without network access.

## License

MIT
