package main

import (
	"ssm/internal/blockchain"
	"ssm/internal/ws"
)

func main() {
	servers := blockchain.NewBlockchain(1)
	servers.AddBlock("127.1.1.1")
	servers.AddBlock("8.8.8.8")

	server := ws.NewServer()
	if err := server.ListenAndServe(3000); err != nil {
		panic(err)
	}
}
