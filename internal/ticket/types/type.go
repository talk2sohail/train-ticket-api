package types

import (
	"context"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

type TicketService interface {
	// Submits a purchase for a train ticket.
	PurchaseTicket(context.Context, *ticket.PurchaseTicketRequest) (ticket.PurchaseTicketResponse, error)
}
