package client

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

// TicketClient wraps the gRPC client and connection.
// It provides methods to interact with the Train Ticketing Service.
// This client is designed to be used in a microservices architecture where
// the ticketing service is a separate gRPC server.
// The client handles the connection to the server and provides methods
// to call the service's RPC methods.
// It is important to ensure that the connection is closed properly
// when the client is no longer needed to avoid resource leaks.
// The client methods log errors and return them to the caller.
// This allows the caller to handle errors appropriately.

// NOTE: The client is currently tighly coupled with grpc code. Need to refactor it to be more generic and decoupled from grpc.
type TicketClient struct {
	conn   *grpc.ClientConn
	client ticket.TrainTicketingServiceClient
}

// NewTrainTicketClient creates a new TicketClient and connects to the given gRPC server address.
func NewTrainTicketClient(serverAddr string) (*TicketClient, error) {
	// In production, consider using secure connections.
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	client := ticket.NewTrainTicketingServiceClient(conn)
	return &TicketClient{conn: conn, client: client}, nil
}

// Close shuts down the gRPC connection.
func (tc *TicketClient) Close() error {
	return tc.conn.Close()
}

// PurchaseTicket forwards the call to the gRPC service.
func (tc *TicketClient) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (*ticket.PurchaseTicketResponse, error) {
	resp, err := tc.client.PurchaseTicket(ctx, req)
	if err != nil {
		log.Printf("PurchaseTicket error: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetReceiptDetails forwards the call to the gRPC service.
func (tc *TicketClient) GetReceiptDetails(ctx context.Context, ticketID string) (*ticket.GetReceiptDetailsResponse, error) {
	req := &ticket.GetReceiptDetailsRequest{Identifier: &ticket.GetReceiptDetailsRequest_TicketId{TicketId: ticketID}}
	resp, err := tc.client.GetReceiptDetails(ctx, req)
	if err != nil {
		log.Printf("GetReceiptDetails error for ticketID %s: %v", ticketID, err)
		return nil, err
	}
	return resp, nil
}

// GetUsersBySection forwards the call to the gRPC service.
func (tc *TicketClient) GetUsersBySection(ctx context.Context, section ticket.Seat_Section) (*ticket.GetUsersBySectionResponse, error) {
	req := &ticket.GetUsersBySectionRequest{Section: section}
	resp, err := tc.client.GetUsersBySection(ctx, req)
	if err != nil {
		log.Printf("GetUsersBySection error for section %s: %v", section.String(), err)
		return nil, err
	}
	return resp, nil
}

// RemoveUser forwards the call to the gRPC service.
func (tc *TicketClient) RemoveUser(ctx context.Context, email string) (*ticket.RemoveUserResponse, error) {
	req := &ticket.RemoveUserRequest{Identifier: &ticket.RemoveUserRequest_Email{Email: email}}
	resp, err := tc.client.RemoveUser(ctx, req)
	if err != nil {
		log.Printf("RemoveUser error for email %s: %v", email, err)
		return nil, err
	}
	return resp, nil
}

// ModifyUserSeat forwards the call to the gRPC service.
func (tc *TicketClient) ModifyUserSeat(ctx context.Context, ticketID string, newSeat *ticket.Seat) (*ticket.ModifyUserSeatResponse, error) {
	req := &ticket.ModifyUserSeatRequest{
		Identifier: &ticket.ModifyUserSeatRequest_TicketId{TicketId: ticketID},
		NewSeat:    newSeat,
	}
	resp, err := tc.client.ModifyUserSeat(ctx, req)
	if err != nil {
		log.Printf("ModifyUserSeat error for ticketID %s: %v", ticketID, err)
		return nil, err
	}
	return resp, nil
}
