# kvault

Key-value store with REST API. Persists data to a JSON file. Supports GET, POST, DELETE operations with concurrent-safe access.

## Install

```bash
cd kvault
go build -o kvault
mv kvault ~/.local/bin/
```

## Usage

```bash
# start with defaults (port 8080, kvault.json)
kvault

# custom port and data file
kvault -port 9090 -data /var/lib/kvault/data.json

# no color
kvault --no-color
```

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/keys` | GET | List all key-value pairs |
| `/keys/{key}` | GET | Get value by key |
| `/keys/{key}` | POST | Set value `{"value":"..."}` |
| `/keys/{key}` | DELETE | Delete key |

## Examples

```bash
# set a value
curl -X POST localhost:8080/keys/secret -d '{"value":"abc123"}'
{"key":"secret","value":"abc123"}

# get a value
curl localhost:8080/keys/secret
{"key":"secret","value":"abc123"}

# list all
curl localhost:8080/keys
{"secret":"abc123"}

# delete
curl -X DELETE localhost:8080/keys/secret
{"key":"secret","deleted":"true"}
```

## Flags

```
-port string    listen port (default "8080")
-data string    path to data file (default "kvault.json")
--no-color      disable colored output
```

## Testing

```bash
go test -v ./...
```

Tests use an in-memory store mock and verify file persistence separately.

## License

MIT
