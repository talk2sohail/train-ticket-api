package server

import (
	"log"
	"net"

	"github.com/talk2sohail/train-ticket-api/internal/ticket/handler"
	"github.com/talk2sohail/train-ticket-api/internal/ticket/service"
	"google.golang.org/grpc"
)

type TicketGRPCServer struct {
	addr string
}

func NewTicketGRPCServer(addr string) *TicketGRPCServer {
	return &TicketGRPCServer{addr: addr}
}

func (s *TicketGRPCServer) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// register our grpc services
	ticketService := service.NewTicketService()
	handler.RegisterTicketServiceServer(grpcServer, ticketService)

	log.Println("Starting Ticketing gRPC server on", s.addr)

	return grpcServer.Serve(lis)
}
