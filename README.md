# Smart Card Reader

Local smart card service written in Go. It exposes the current card state over HTTP on `http://localhost:4121/card`.

Running the built binary directly starts it in background mode by default.

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
go run ./cmd/card-reader serve
```

Background mode:

```bash
make start
make status
make stop
```

Logs and PID file are written into `runtime/`.

If you open the built binary by double-clicking it, it will detach and continue running in background. Use `make status` to check it and `make stop` to stop it.

## Windows startup

On Windows, the first successful `start` automatically registers the app under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`. That means after the user logs in again, Windows starts the app automatically.

Manual commands:

```bash
make startup-on
make startup-status
make startup-off
```

Runtime files are not tied to the current working directory. They are stored in the user's config directory, so reboot and startup launches use the same PID and log location.

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
