# go-fordpass

Ford Pass is Ford's connected vehicle platform. On select vehicles, this enables first-party app control of locks, remote start, and vehicle telematics.

This project is a (partial) reimplementation of [ffpass](https://github.com/d4v3y0rk/ffpass), in Go.

# Features

On supported vehicles, the CLI supports

- Fetch Status (not yet pretty 💩)
- Lock/Unlock Doors
- Start/Stop Engine (Remote Start)

# CLI Usage

### `go get` or Clone Repo:

```
go get -u github.com/parrotmac/go-fordpass
# - or -
git clone https://github.com/parrotmac/go-fordpass.git
```

### Get Credentials
Ensure environment variables are setup, similar to below (place in a `.env` file and `source` it to keep things tidy):

```
export FORD_USERNAME=you@example.com
export FORD_PASSWORD=secret42
export VEHICLE_VIN=1FXXXXXXXXXXXXXXX
```

Alternatively, if not found in the environment, the CLI will prompt for the above values.

### Run

From the project directory, run

```
$ go run cmd/cli.go
Use the arrow keys to navigate: ↓ ↑ → ←
? Choose Action:
  ▸ Get Status
    Lock Doors
    Unlock Doors
    Start Engine
    Stop Engine
```
