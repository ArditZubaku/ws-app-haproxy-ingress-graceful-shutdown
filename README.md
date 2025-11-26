# HAProxy Ingress Graceful Shutdown Testing

A comprehensive testing environment for validating HAProxy Ingress graceful shutdown behavior with WebSocket connections in Kubernetes.

## What We're Testing

This project tests **graceful shutdown scenarios** where:

- HAProxy Ingress Controller receives termination signals
- Active WebSocket connections should be properly closed
- Long-running operations should be interrupted gracefully
- Clients should receive proper close notifications
- No connections should be abruptly dropped

## Architecture

### Components

1. **WebSocket Server (Go)** - `go/cmd/ws_server/`

   - HTTP server on port 8080 (WebSocket endpoint + health checks)
   - TCP service communication server on port 9999
   - Manages active WebSocket connections
   - Handles graceful shutdown with connection cleanup

2. **Cleanup Service (Go)** - `go/cmd/cleanup_svc/`

   - Separate service that communicates with WebSocket server
   - Sends commands to close specific number of connections
   - Simulates external control plane operations

3. **WebSocket Clients (Node.js)** - `nodejs/`

   - Configurable number of persistent WebSocket connections
   - Supports retry logic and connection monitoring
   - Two modes: regular ping/pong and slow operations

4. **HAProxy Ingress Controller**
   - Kubernetes ingress controller handling WebSocket traffic
   - Configured with extended termination grace period (901s)
   - Routes traffic from external clients to WebSocket server

### Communication Flow

```
Node.js Clients → HAProxy Ingress → WebSocket Server
                                          ↕
                                   Cleanup Service
```

## Setup and Testing

### Prerequisites

- Docker
- Minikube
- kubectl
- Helm
- Node.js

### Automated Setup

Run the complete setup script:

```bash
go run setup.go
```

This will:

1. Clean up existing resources
2. Start Minikube cluster
3. Build and load Docker images
4. Deploy HAProxy Ingress Controller with custom configuration
5. Deploy WebSocket server
6. Start 100 Node.js WebSocket clients
7. Deploy cleanup service

### Manual Testing Steps

#### 1. Monitor WebSocket Connections

```bash
# Watch WebSocket server logs
kubectl logs -f deployment/ws-app

# Watch HAProxy controller logs
kubectl logs -f -n haproxy-controller deployment/haproxy-ingress-kubernetes-ingress
```

#### 2. Monitor WebSocket Client Logs

```bash
# View client connection logs
tail -f ws-clients.log
```

#### 3. Test Connection Cleanup

```bash
# Watch cleanup service logs
kubectl logs -f deployment/cleanup-svc
```

The cleanup service automatically sends "11" every 10 seconds, requesting the server to close 11 WebSocket connections.

#### 4. Test Graceful Shutdown

**Scenario A: HAProxy Controller Restart**

```bash
# Restart HAProxy controller to test graceful shutdown
kubectl rollout restart deployment/haproxy-ingress-kubernetes-ingress -n haproxy-controller

# Monitor connection behavior
kubectl logs -f deployment/ws-app
```

**Scenario B: WebSocket Server Restart**

```bash
# Restart WebSocket server
kubectl rollout restart deployment/ws-app

# Monitor client reconnection behavior
tail -f ws-clients.log
```

**Scenario C: Manual Connection Cleanup**

```bash
# Scale down cleanup service and send manual commands
kubectl scale deployment cleanup-svc --replicas=0

# Connect directly to service communication port
kubectl port-forward service/ws-app 9999:9999

# In another terminal, send cleanup commands
echo "50" | nc localhost 9999  # Close 50 connections
```

### Client Configuration

The Node.js clients support multiple configuration options:

```bash
# Basic usage - 1 client
node nodejs/ws.mjs

# Multiple clients
node nodejs/ws.mjs -n 100

# Enable automatic retries on connection failure
node nodejs/ws.mjs -n 100 -r

# Use slow endpoint mode (30-second operations)
node nodejs/ws.mjs -n 50 -s

# Show help
node nodejs/ws.mjs -h
```

## Expected Results

### Successful Graceful Shutdown

**WebSocket Server Logs:**

```
INFO Service communication server listening on addr=[::]:9999
INFO WebSocket connection added total=100
INFO Received service message message=11
INFO Closing WebSocket connections count=11
INFO WebSocket connection removed total=89
```

**Client Logs (with retries enabled):**

```
[Client 1] Connected to ws://haproxy.local:30234
[Client 1][ws://haproxy.local:30234][1.2s] Echo: Hello from Node.js client!
[Client 1][ws://haproxy.local:30234] Connection closed after 45.3s
[Client 1][ws://haproxy.local:30234] Close code: 1001, reason: Server shutting down
[Client 1] Connection dropped. Retrying in 1000ms...
[Client 1][Attempt 2/6] Connecting to ws://haproxy.local:30234...
```

**HAProxy Controller Logs:**

```
[NOTICE] haproxy version X.X.X-XXXX
[WARNING] stopping frontend haproxy-ingress-kubernetes-ingress_http
[NOTICE] graceful stop of proxy haproxy-ingress-kubernetes-ingress_http
```

### Connection State Verification

Monitor connection counts:

```bash
# Check active connections via WebSocket server
kubectl exec deployment/ws-app -- ss -tln | grep :8080

# Check HAProxy stats (if enabled)
kubectl exec -n haproxy-controller deployment/haproxy-ingress-kubernetes-ingress -- echo "show stat" | socat stdio /var/run/haproxy.sock
```

## Troubleshooting

### Common Issues

1. **Clients can't connect**

   ```bash
   # Check ingress configuration
   kubectl get ingress haproxy-ingress
   kubectl describe ingress haproxy-ingress

   # Verify service endpoints
   kubectl get endpoints ws-app
   ```

2. **Service communication fails**

   ```bash
   # Check if TCP port 9999 is accessible
   kubectl port-forward service/ws-app 9999:9999
   telnet localhost 9999
   ```

3. **HAProxy not routing WebSocket traffic**
   ```bash
   # Check HAProxy configuration
   kubectl exec -n haproxy-controller deployment/haproxy-ingress-kubernetes-ingress -- cat /etc/haproxy/haproxy.cfg | grep -A 10 -B 10 websocket
   ```

### Debug Commands

```bash
# Show all resources
kubectl get all,ingress,configmap,secret

# Show HAProxy controller configuration
kubectl get configmap -n haproxy-controller

# Test direct connection to WebSocket server (bypass HAProxy)
kubectl port-forward service/ws-app 8080:8080
# Then connect to ws://localhost:8080

# View all logs
kubectl logs deployment/ws-app
kubectl logs deployment/cleanup-svc
kubectl logs -n haproxy-controller deployment/haproxy-ingress-kubernetes-ingress
```

## Testing Scenarios

### Scenario 1: Normal Operation

- Deploy all components
- Verify 100 clients connect successfully
- Observe cleanup service closing connections periodically
- Expected: Stable connections with graceful cleanup

### Scenario 2: HAProxy Controller Restart

- Trigger HAProxy controller restart
- Monitor connection state during transition
- Expected: Connections gracefully transferred to new HAProxy instance

### Scenario 3: WebSocket Server Restart

- Restart WebSocket server pod
- Monitor client reconnection behavior
- Expected: Clients detect disconnection and reconnect

### Scenario 4: Network Partition Simulation

- Block traffic between HAProxy and WebSocket server
- Monitor timeout and recovery behavior
- Expected: Proper error handling and connection cleanup

### Scenario 5: High Load Testing

- Scale clients to 1000+ connections
- Test performance during graceful shutdown
- Expected: All connections properly closed without timeout

## Configuration Details

### HAProxy Controller Configuration

- **Termination Grace Period**: 901 seconds (extended for testing)
- **Service Type**: NodePort (for external access)
- **Image**: haproxytech/kubernetes-ingress:3.1.14

### WebSocket Server Configuration

- **HTTP Port**: 8080 (WebSocket + health endpoints)
- **Service Communication Port**: 9999 (TCP)
- **Health Check Endpoints**: `/healthz`
- **Connection Timeout**: Configurable via environment

### Kubernetes Resources

- **Namespace**: default (WebSocket server, cleanup service)
- **Namespace**: haproxy-controller (HAProxy Ingress Controller)
- **Service Type**: ClusterIP (internal), NodePort (HAProxy)

This testing environment provides comprehensive validation of graceful shutdown behavior in a realistic Kubernetes deployment with HAProxy Ingress.
