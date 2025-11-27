package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"time"
)

func main() {
	const serviceEndpoint = "ws-app:9999"

	conn, err := net.Dial("tcp", serviceEndpoint)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	slog.Info("Connected to service at ", "addr", conn.RemoteAddr().String())

	for {
		buf := make([]byte, 1)
		n, err := conn.Read(buf)
		if err != nil {
			slog.Error("Failed to read from service", "error", err)
			return
		}

		slog.Info(
			"Received message from server - that means 100 clients have been connected",
			"message", string(buf[:n]),
		)

		if n > 0 {
			break
		}
	}

	scanner := bufio.NewScanner(conn)
	for {
		n, err := fmt.Fprintln(conn, "11")
		if err != nil || n == 0 {
			slog.Error("Failed to write to service", "error", err)
			return
		}

		if scanner.Scan() {
			slog.Info("Received from service", "message", scanner.Text())
		}

		time.Sleep(10 * time.Second)
	}
}
