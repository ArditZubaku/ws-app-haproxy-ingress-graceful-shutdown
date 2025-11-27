// Package tcp provides inter-service communication functionality via TCP.
package tcp

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/ArditZubaku/go-node-ws/internal/connmanager"
)

func HandleServiceCommunication(
	cm *connmanager.ConnectionManager,
	sendToCleanupSvc <-chan struct{},
) {
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		slog.Error("Failed to listen on TCP port", "error", err)
		return
	}
	defer ln.Close()

	slog.Info("Service communication server listening on", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Failed to accept IPC connection", "error", err)
			continue
		}
		go handleServiceConnection(conn, cm, sendToCleanupSvc)
	}
}

func handleServiceConnection(
	conn net.Conn,
	cm *connmanager.ConnectionManager,
	sendToCleanupSvc <-chan struct{},
) {
	defer conn.Close()

	<-sendToCleanupSvc
	conn.Write([]byte{1}) // Notify cleanup service to start - just a byte signal

	reader := bufio.NewScanner(conn)

	for reader.Scan() {
		msg := reader.Text()
		n, err := strconv.Atoi(msg)
		if err != nil {
			slog.Error("Invalid number received", "error", err)
			continue
		}
		slog.Info("Received service message", "message", n)

		cm.CloseNConnections(n)

		// No need for newline, fmt.Fprintln adds it
		n, err = fmt.Fprintln(conn, "Closing "+msg+" WS connections")
		if n == 0 || err != nil {
			slog.Error("Failed to write service response", "error", err)
			return
		}
	}
}
