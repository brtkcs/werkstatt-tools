# netmapper

Network host discovery and port scanner for /24 subnets. Scans 254 hosts concurrently using goroutines, then checks open ports on each live host.

## Install

```bash
cd netmapper
go build -o netmapper
mv netmapper ~/.local/bin/
```

## Usage

```bash
# scan default subnet
netmapper -subnet 192.168.1

# custom timeout and ports
netmapper -subnet 10.0.0 -timeout 1000 -ports 22,80,443,8080

# no color (for logging/pipes)
netmapper --no-color
```

## How it works

Two-phase scan:

1. **Host discovery** - 254 goroutines probe the subnet simultaneously, checking common ports on each IP to determine if the host is alive.
2. **Port scan** - each live host is scanned concurrently for open ports using goroutines with mutex-protected results.

## Default ports

22, 53, 80, 443, 631, 3000, 3306, 5432, 8080, 8443, 9090

## Example output

```
netmapper -> 10.0.0.0/24

discovering hosts...
4 hosts found, scanning ports...

  * 10.0.0.1       22  80  443
  * 10.0.0.10      22  8080
  * 10.0.0.20      22  3306  5432
  * 10.0.0.50      80  443

  4 hosts
```

## Flags

```
-subnet string    subnet prefix, first 3 octets (default "192.168.1")
-timeout int      connection timeout in ms (default 500)
-ports string     comma-separated port list (default: common ports)
--no-color        disable colored output
```

## Testing

```bash
go test -v ./...
```

Tests use local TCP listeners and httptest servers to verify scanning logic without touching the network.

## License

MIT
