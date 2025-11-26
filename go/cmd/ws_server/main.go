package main

import (
	"log/slog"

	"github.com/ArditZubaku/go-node-ws/internal/connmanager"
	"github.com/ArditZubaku/go-node-ws/internal/ipc"
	"github.com/ArditZubaku/go-node-ws/internal/server"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	cm := connmanager.NewConnectionManager()
	go ipc.HandleIPCCommunication(cm)
	server.NewServer(cm).Start()
}
