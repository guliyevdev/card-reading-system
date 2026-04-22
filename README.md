# Smart Card Reader

Local smart card service written in Go. It exposes the current card state over HTTP on `http://localhost:4121/card`.

## Requirements

- Go 1.26 or newer
- A working PC/SC stack on the host OS
- Supported reader attached to the machine

## Run

```bash
make run
```

Or directly:

```bash
go run ./cmd/card-reader
```

## Build

Native build:

```bash
make build
```

Platform builds:

```bash
make build-mac-arm64
make build-win-x64
```

## API

```http
GET /card
```

Response:

```json
{
  "uid": "04A1B2C3D4",
  "atr": "3B8F8001804F0CA000000306030001000000006A"
}
```

When no card is present:

```json
{
  "uid": null,
  "atr": null
}
```
