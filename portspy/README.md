# portspy

Concurrent TCP port scanner. Scans a port range on a target host using goroutines, reporting open ports with response times.

## Install

```bash
cd portspy
go build -o portspy
mv portspy ~/.local/bin/
```

## Usage

```bash
# scan localhost, ports 1-1024
portspy

# scan a remote host
portspy -host 10.0.0.1

# full range with longer timeout
portspy -host 192.168.1.1 -start 1 -end 65535 -timeout 2

# no color (for logging/pipes)
portspy --no-color
```

## Example output

```
portspy -> 10.0.0.1 [1-1024]

  22     open  12ms
  80     open  8ms
  443    open  9ms

  3 open ports
```

## Flags

```
-host string    target host or IP (default "localhost")
-start int      start port (default 1)
-end int        end port (default 1024)
-timeout int    timeout in seconds (default 1)
--no-color      disable colored output
```

## Testing

```bash
go test -v ./...
```

Tests use local TCP listeners to verify scanning logic without touching the network.

## License

MIT
