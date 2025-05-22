package types

import (
	"github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

type TicketService interface {
	// Submits a purchase for a train ticket.
	PurchaseTicket(request ticket.PurchaseTicketRequest) (ticket.PurchaseTicketResponse, error)

	// Retrieves the details of a specific receipt for a user.
	GetReceiptDetails(request ticket.GetReceiptDetailsRequest) (ticket.GetReceiptDetailsResponse, error)

	// Views all users and their allocated seats for a given train section.
	GetUsersBySection(request ticket.GetUsersBySectionRequest) (ticket.GetUsersBySectionResponse, error)

	// Removes a user and their ticket from the train.
	RemoveUser(request ticket.RemoveUserRequest) (ticket.RemoveUserResponse, error)

	// Modifies the seat allocation for an existing user.
	ModifyUserSeat(request ticket.ModifyUserSeatRequest) (ticket.ModifyUserSeatResponse, error)
}
