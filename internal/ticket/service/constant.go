package service

const (
	// MaxSeatsPerSection defines the maximum number of seats in each section.
	MaxSeatsPerSection = 5

	// useful message
	MsgTicketPurchaseSuccess = "Ticket purchased successfully"
	MsgUsersRetrieved        = "Users retrieved successfully"
	MsgUserRemovedSuccess    = "User removed successfully"
	MsgSeatUpdatedSuccess    = "Seat updated successfully"

	// Define named errors
	ErrNoAvailableSeats = "no available seats on the train"
	ErrReceiptNotFound  = "receipt not found"
	ErrUserNotFound     = "user not found"
	ErrSeatOccupied     = "requested seat is already occupied"
)
