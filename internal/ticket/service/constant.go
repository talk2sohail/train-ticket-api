package service

const (
	// MaxSeatsPerSection defines the maximum number of seats in each section.
	MaxSeatsPerSection = 5

	// Define named errors
	ErrNoAvailableSeats = "no available seats on the train"
	ErrReceiptNotFound  = "receipt not found"
	ErrUserNotFound     = "user not found"
	ErrSeatOccupied     = "requested seat is already occupied"
)
