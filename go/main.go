package main

import (
	"github.com/ArditZubaku/go-node-ws/internal/conn_manager"
	"github.com/ArditZubaku/go-node-ws/internal/server"
)

func main() {
	server.NewServer(conn_manager.NewConnectionManager()).Start()
}
