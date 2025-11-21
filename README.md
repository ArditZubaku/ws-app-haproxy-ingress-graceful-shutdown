# WebSocket & HTTP Graceful Shutdown Testing

A Go WebSocket server with Node.js client for testing graceful shutdown behavior

## Features

- **WebSocket Server (Go)**: Handles persistent WebSocket connections with echo responses
- **HTTP Slow Endpoint**: Simulates long-running operations (30 seconds)
- **Graceful Shutdown**: Properly closes WebSocket connections and interrupts long operations on SIGINT/SIGTERM
- **WebSocket Client (Node.js)**: Connects persistently and sends messages every 3 seconds
- **Slow Operation Mode**: Test graceful shutdown with long-running WebSocket operations

## Quick Start

### 1. Start the Go Server

```bash
go run main.go
```

The server will start on `localhost:8080` with:

- WebSocket endpoint: `ws://localhost:8080/`

### 2. Run the Node.js Client

**Regular mode** (normal WebSocket communication):

```bash
node ws.mjs
```

**Slow mode** (triggers 30-second operations via WebSocket):

```bash
node ws.mjs -s
```

## Testing Graceful Shutdown

1. Start the server: `go run main.go`
2. Start the client: `node ws.mjs` or `node ws.mjs -s`
3. Send `SIGINT` (Ctrl+C) or `SIGTERM` to the server
4. Observe how:
   - WebSocket connections are gracefully closed
   - Long-running operations are interrupted
   - Clients receive proper close notifications

## Endpoints

### WebSocket: `ws://localhost:8080/`

- Accepts persistent connections
- Echoes all messages back to client
- Special handling for `SLOW_REQUEST` and `SLOW_PING` messages (30s operations)

## Example Output

**Server:**

```
INFO Starting server address=:8080
INFO WebSocket connection added total=1
INFO Received message message="Hello from Node.js client!"
INFO Sent echo back to client
```

**Client:**

```
Connected to ws://127.0.0.1:8080
[ws://127.0.0.1:8080][0.0s] WebSocket connection established
[ws://127.0.0.1:8080][0.0s] Echo: Hello from Node.js client!
[ws://127.0.0.1:8080][3.0s] Echo: Ping from client at 2025-11-21T10:34:45.513Z
```
