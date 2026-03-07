# dmon

Process monitor with state change detection. Watches processes from a YAML config and reports when they start or stop.

## Install

```bash
cd dmon
go build -o dmon
mv dmon ~/.local/bin/
```

Requires `gopkg.in/yaml.v3`:

```bash
go mod tidy
```

## Usage

```bash
# continuous monitoring (default)
dmon

# run once and exit (for cron/scripts)
dmon -once

# custom config
dmon -c /etc/dmon/processes.yaml

# no color (for logging)
dmon --no-color
```

## Config

Create `dmon.yaml`:

```yaml
interval: 5
processes:
  - sshd
  - podman
  - nvim
  - firefox
```

The `interval` is in seconds (default: 5).

## Example output

First run shows current state:

```
dmon -> 4 processes (5s)

  * sshd              running
  * podman            running
  * nvim              running
  * firefox           not running
```

When a process stops or starts:

```
  14:22:05  v nvim  stopped
  14:23:10  ^ nvim  started
```

## Flags

```
-c string       path to config file (default "dmon.yaml")
-once           run once and exit
--no-color      disable colored output
```

## Testing

```bash
go test -v ./...
```

Tests use a mock checker interface to simulate process states without calling pgrep.

## License

MIT
