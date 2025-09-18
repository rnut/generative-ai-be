# Workshop BE (Fiber)

Simple Go Fiber starter.

## Prerequisites
- Go 1.22+

## Run
```bash
go run main.go
```

Then open: http://localhost:3000

Endpoints:
- `/` returns JSON `{"message":"Hello Fiber"}`
- `/healthz` returns 200

## Hot Reload (Air)
Install once (adds to GOPATH bin):
```bash
go install github.com/air-verse/air@latest
```
Run with config:
```bash
air
```
หรือกำหนดไฟล์กำหนดเอง:
```bash
air -c .air.toml
```

## Build
```bash
go build -o workshop-be
```

## Environment Variables
- `PORT` (default 3000)

## Graceful Shutdown
Uses `app.Shutdown()` when receiving SIGINT/SIGTERM.
