package handler

import (
	"context"
	"errors"
	"log"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
	"github.com/talk2sohail/train-ticket-api/internal/common/util"
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

	// Validate the request object.
	err := util.ValidatePurchseRequestObject(req)
	if err != nil {
		log.Printf("Invalid PurchaseTicket request: %v", err)
		return nil, err
	}

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

// GetReceiptDetails handles the retrieval of receipt details for a given ticket.
func (h *TicketGrpcHandler) GetReceiptDetails(ctx context.Context, req *ticket.GetReceiptDetailsRequest) (*ticket.GetReceiptDetailsResponse, error) {
	// Optionally, validate request here if needed.
	receipt, err := h.ticketService.GetReceiptDetails(ctx, req.GetTicketId())
	if err != nil {
		log.Printf("Error retrieving receipt for ticketID %s: %v", req.GetTicketId(), err)
		return nil, err
	}
	return &ticket.GetReceiptDetailsResponse{
		Receipt: receipt,
	}, nil
}

// GetUsersBySection handles the retrieval of users and their seats by section.
func (h *TicketGrpcHandler) GetUsersBySection(ctx context.Context, req *ticket.GetUsersBySectionRequest) (*ticket.GetUsersBySectionResponse, error) {
	// Validate the section.
	err := util.ValidateSection(req)
	if err != nil {
		log.Printf("Invalid section in GetUsersBySection request: %v", err)
		return nil, err
	}

	resp, err := h.ticketService.GetUsersBySection(ctx, req.GetSection())
	if err != nil {
		log.Printf("Error in GetUsersBySection: %v", err)
		return nil, err
	}
	return &resp, nil
}

// RemoveUser handles removing a user from the train.
func (h *TicketGrpcHandler) RemoveUser(ctx context.Context, req *ticket.RemoveUserRequest) (*ticket.RemoveUserResponse, error) {
	email := req.GetEmail()
	if email == "" {
		return nil, errors.New("email is required")
	}
	resp, err := h.ticketService.RemoveUser(ctx, email)
	if err != nil {
		log.Printf("Error in RemoveUser: %v", err)
		return nil, err
	}
	return &resp, nil
}

// ModifyUserSeat handles updating a user's seat allocation.
func (h *TicketGrpcHandler) ModifyUserSeat(ctx context.Context, req *ticket.ModifyUserSeatRequest) (*ticket.ModifyUserSeatResponse, error) {
	// Validate identifier: ensure ticketId is provided.
	ticketID := req.GetTicketId()
	if ticketID == "" {
		return nil, errors.New("ticketId is required")
	}
	// Retrieve existing receipt.
	receipt, err := h.ticketService.GetReceiptDetails(ctx, ticketID)
	if err != nil {
		log.Printf("Error retrieving receipt for ticketID %s: %v", ticketID, err)
		return nil, err
	}
	// Ensure new seat is provided.
	newSeat := req.GetNewSeat()
	if newSeat == nil {
		return nil, errors.New("new seat is required")
	}

	// Call the service method with the receipt and new seat.
	resp, err := h.ticketService.ModifyUserSeat(ctx, receipt, newSeat)
	if err != nil {
		log.Printf("Error in ModifyUserSeat: %v", err)
		return nil, err
	}
	return &resp, nil
}
