package main

import (
	"log/slog"

	"github.com/ArditZubaku/go-node-ws/internal/connmanager"
	"github.com/ArditZubaku/go-node-ws/internal/http"
	"github.com/ArditZubaku/go-node-ws/internal/tcp"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	sendToCleanupSvc := make(chan struct{})
	cm := connmanager.NewConnectionManager(sendToCleanupSvc)
	go tcp.HandleServiceCommunication(cm, sendToCleanupSvc)
	http.NewServer(cm).Start()
}
