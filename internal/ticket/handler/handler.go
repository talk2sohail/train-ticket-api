package handler

import (
	"context"
	"log"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
	"github.com/talk2sohail/train-ticket-api/internal/ticket/types"
	"google.golang.org/grpc"
)

type TicketGrpcHandler struct {
	ticketService types.TicketService
	ticket.UnimplementedTrainTicketingServiceServer
}

func RegisterTicketServiceServer(grpc *grpc.Server, ticketService types.TicketService) {
	gRPCHandler := &TicketGrpcHandler{
		ticketService: ticketService,
	}

	// register the ticket service server
	ticket.RegisterTrainTicketingServiceServer(grpc, gRPCHandler)
}

func (h *TicketGrpcHandler) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (*ticket.PurchaseTicketResponse, error) {
	log.Printf("Received PurchaseTicket request: From=%s, To=%s, User=%s %s (%s), Price=%.2f",
		req.GetFromLocation(), req.GetToLocation(),
		req.GetUser().GetFirstName(), req.GetUser().GetLastName(), req.GetUser().GetEmail(),
		req.GetPricePaid())

	response, err := h.ticketService.PurchaseTicket(ctx, req)
	if err != nil {
		log.Printf("Error processing PurchaseTicket request: %v", err)
		return nil, err
	}
	return &response, nil
}
