package types

import (
	"context"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

type TicketService interface {
	PurchaseTicket(context.Context, *ticket.PurchaseTicketRequest) (ticket.PurchaseTicketResponse, error)
	GetReceiptDetails(context.Context, string) (*ticket.Receipt, error)
	GetUsersBySection(context.Context, ticket.Seat_Section) (ticket.GetUsersBySectionResponse, error)
	RemoveUser(context.Context, string) (ticket.RemoveUserResponse, error)
	ModifyUserSeat(context.Context, *ticket.Receipt, *ticket.Seat) (ticket.ModifyUserSeatResponse, error)
}
