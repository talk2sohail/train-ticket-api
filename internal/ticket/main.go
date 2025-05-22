package main

import "github.com/talk2sohail/train-ticket-api/internal/ticket/server"

func main() {
	grpcServer := server.NewTicketGRPCServer(":9001")
	grpcServer.Run()
}
