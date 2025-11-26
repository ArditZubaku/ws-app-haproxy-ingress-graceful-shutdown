// Package ipc provides inter-process communication functionality via Unix sockets.
package ipc

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/ArditZubaku/go-node-ws/internal/connmanager"
)

func HandleIPCCommunication(cm *connmanager.ConnectionManager) {
	const socketPath = "/tmp/ipc.sock"
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		slog.Error("Failed to listen on unix socket", "error", err)
		return
	}
	defer ln.Close()

	slog.Info("IPC server listening on", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("Failed to accept IPC connection", "error", err)
			continue
		}
		go handleIPCConnection(conn, cm)
	}
}

func handleIPCConnection(conn net.Conn, cm *connmanager.ConnectionManager) {
	defer conn.Close()
	reader := bufio.NewScanner(conn)

	for reader.Scan() {
		msg := reader.Text()
		n, err := strconv.Atoi(msg)
		if err != nil {
			slog.Error("Invalid number received", "error", err)
			continue
		}
		slog.Info("Received IPC message", "message", n)

		cm.CloseNConnections(n)

		// No need for newline, fmt.Fprintln adds it
		n, err = fmt.Fprintln(conn, "Closing "+msg+" WS connections")
		if n == 0 || err != nil {
			slog.Error("Failed to write IPC response", "error", err)
			return
		}
	}
}
