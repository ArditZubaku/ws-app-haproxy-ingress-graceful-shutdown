// Package connmanager provides WebSocket connection management functionality.
// It tracks active connections and handles graceful shutdown procedures.
package connmanager

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionManager tracks and manages WebSocket connections
type ConnectionManager struct {
	connections      map[*websocket.Conn]bool
	mu               sync.RWMutex
	Shutdown         chan struct{}
	SendToCleanupSvc chan<- struct{}
}

func NewConnectionManager(
	sendToCleanupSvc chan<- struct{},
) *ConnectionManager {
	return &ConnectionManager{
		connections:      make(map[*websocket.Conn]bool),
		Shutdown:         make(chan struct{}),
		SendToCleanupSvc: sendToCleanupSvc,
	}
}

func (cm *ConnectionManager) AddConnection(conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.connections[conn] = true
	slog.Info("WebSocket connection added", "total", len(cm.connections))
	if len(cm.connections) >= 100 {
		slog.Info("Reached 100 WebSocket connections")
		close(cm.SendToCleanupSvc)
	}
}

func (cm *ConnectionManager) RemoveConnection(conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, conn)
	slog.Info("WebSocket connection removed", "total", len(cm.connections))
}

func (cm *ConnectionManager) GetNConnections(n int) []*websocket.Conn {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connections := make([]*websocket.Conn, 0, n)
	for conn := range cm.connections {
		if len(connections) >= n {
			break
		}
		connections = append(connections, conn)
	}
	return connections
}

func (cm *ConnectionManager) CloseNConnections(n int) {
	connections := cm.GetNConnections(n)

	slog.Info("Closing WebSocket connections", "count", len(connections))

	for _, conn := range connections {
		// Send close message
		if err := conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseGoingAway,
				"Server shutting down",
			),
		); err != nil {
			slog.Error("Error sending close message", "error", err)
		}

		if err := conn.Close(); err != nil {
			slog.Error("Error closing WebSocket connection", "error", err)
		}

		// Remove from map
		cm.RemoveConnection(conn)
	}
}

func (cm *ConnectionManager) CloseAllConnections(ctx context.Context) {
	cm.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(cm.connections))
	for conn := range cm.connections {
		connections = append(connections, conn)
	}
	cm.mu.RUnlock()

	slog.Info("Closing all WebSocket connections", "count", len(connections))

	// Signal shutdown to all connections
	close(cm.Shutdown)

	// Close all connections gracefully
	for _, conn := range connections {
		// Send close message
		if err := conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseGoingAway,
				"Server shutting down",
			),
		); err != nil {
			slog.Error("Error sending close message", "error", err)
		}

		if err := conn.Close(); err != nil {
			slog.Error("Error closing WebSocket connection", "error", err)
		}
	}

	// Wait for all connections to be removed or timeout
	timeout := time.NewTimer(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Warn("Context cancelled while waiting for WebSocket connections to close")
			return
		case <-timeout.C:
			slog.Warn("Timeout waiting for WebSocket connections to close")
			return
		case <-ticker.C:
			cm.mu.RLock()
			count := len(cm.connections)
			cm.mu.RUnlock()
			if count == 0 {
				slog.Info("All WebSocket connections closed")
				return
			}
		}
	}
}
