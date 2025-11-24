package main

import (
	"log/slog"

	"github.com/ArditZubaku/go-node-ws/internal/conn_manager"
	"github.com/ArditZubaku/go-node-ws/internal/server"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	server.NewServer(conn_manager.NewConnectionManager()).Start()
}
