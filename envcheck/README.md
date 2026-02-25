# envcheck

Validate `.env` files against `.env.example`. Catches missing keys, duplicates, empty values, and syntax errors.

## Install

```bash
cd envcheck
go build -o envcheck
# optional: move to PATH
mv envcheck ~/.local/bin/
```

## Usage

```bash
# check .env against .env.example in current directory
envcheck

# specify files
envcheck -f config/.env -e config/.env.example

# quiet mode: only show errors (useful in CI)
envcheck -q

# no color output (for logging/pipes)
envcheck --no-color
```

## What it checks

| Check | Level | Example |
|-------|-------|---------|
| Missing keys | error | key in `.env.example` but not in `.env` |
| Duplicate keys | error | same key appears twice in `.env` |
| Invalid syntax | error | line without `=` sign |
| Empty values | warning | `DB_PASSWORD=` |
| Extra keys | warning | key in `.env` but not in `.env.example` |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | no errors (warnings are ok) |
| 1 | validation errors found |
| 2 | file not found |

## Example output

```
 envcheck
──────────────────────────────────
✗ :7 invalid syntax: INVALID LINE
✗ :6 duplicate key: DUPLICATE (first: line 5)
⚠ :4 empty value: DB_PASSWORD
✗ missing key: DB_NAME (defined in .env.example)
✗ missing key: DB_USER (defined in .env.example)
✗ missing key: APP_PORT (defined in .env.example)
✗ missing key: APP_ENV (defined in .env.example)
✗ missing key: APP_SECRET (defined in .env.example)
✗ missing key: REDIS_URL (defined in .env.example)

──────────────────────────────────
5 keys  8 errors  1 warnings
```

## Flags

```
-f string    path to .env file (default ".env")
-e string    path to .env.example file (default ".env.example")
-q           quiet mode: only show errors
--no-color   disable colored output
```

## Testing

```bash
go test -v ./...
```

## License

MIT
