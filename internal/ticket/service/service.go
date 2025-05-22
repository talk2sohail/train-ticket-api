package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TicketService struct {
	mu                sync.Mutex                  // Mutex to protect concurrent access to in-memory data structures.
	receipts          map[string]*ticket.Receipt  // Stores all purchased receipts, keyed by Ticket ID.
	occupiedSeats     map[string]*ticket.Receipt  // Stores which seats are occupied, keyed by seat number (e.g., "A1").
	sectionCapacities map[ticket.Seat_Section]int // Defines the maximum number of seats for each section.
}

// NewTicketService creates a new instance of TicketService
func NewTicketService() *TicketService {
	return &TicketService{
		// Initialize any necessary fields here
		receipts:      make(map[string]*ticket.Receipt), // Initialize the map for receipts.
		occupiedSeats: make(map[string]*ticket.Receipt), // Initialize the map for occupied seats.
		sectionCapacities: map[ticket.Seat_Section]int{ // Define the capacity for each section.
			ticket.Seat_SECTION_A: 5, // Section A has 5 seats (A1 to A5).
			ticket.Seat_SECTION_B: 5, // Section B has 5 seats (B1 to B5).
		},
	}
}

// findNextAvailableSeat iterates through sections and seat numbers to find the first unoccupied seat.
// This function assumes the caller has already acquired the server's mutex to ensure thread safety
// when accessing `s.occupiedSeats` and `s.sectionCapacities`.
func (s *TicketService) findNextAvailableSeat() (*ticket.Seat, error) {
	// First, attempt to find an available seat in Section A.
	for i := 1; i <= s.sectionCapacities[ticket.Seat_SECTION_A]; i++ {
		seatNumber := fmt.Sprintf("A%d", i) // Construct seat string, e.g., "A1", "A2"
		// Check if the seat is already occupied.
		if _, isOccupied := s.occupiedSeats[seatNumber]; !isOccupied {
			// If not occupied, return this seat.
			return &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: seatNumber,
			}, nil
		}
	}

	// If Section A is full, attempt to find an available seat in Section B.
	for i := 1; i <= s.sectionCapacities[ticket.Seat_SECTION_B]; i++ {
		seatNumber := fmt.Sprintf("B%d", i) // Construct seat string, e.g., "B1", "B2"
		// Check if the seat is already occupied.
		if _, isOccupied := s.occupiedSeats[seatNumber]; !isOccupied {
			// If not occupied, return this seat.
			return &ticket.Seat{
				Section:    ticket.Seat_SECTION_B,
				SeatNumber: seatNumber,
			}, nil
		}
	}

	// If both sections are fully occupied, return an error.
	return nil, fmt.Errorf("no available seats on the train")
}

// PurchaseTicket handles the purchase of a train ticket
func (s *TicketService) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (ticket.PurchaseTicketResponse, error) {

	// Acquire a lock to protect shared server state (receipts, occupiedSeats) during concurrent access.
	s.mu.Lock()
	defer s.mu.Unlock() // Ensure the lock is released when the function exits.

	// Attempt to find the next available seat using our allocation logic.
	allocatedSeat, err := s.findNextAvailableSeat()
	if err != nil {
		// If no seats are available, log the error and return a failed response.
		log.Printf("Failed to purchase ticket for user %s: %v", req.GetUser().GetEmail(), err)
		return ticket.PurchaseTicketResponse{
			Success: false,
			Message: err.Error(), // Provide the reason for failure.
			Receipt: nil,         // No receipt generated on failure.
		}, nil
	}

	// Generate a unique ticket ID for the new purchase.
	ticketID := uuid.New().String()

	// Construct the Receipt object using the request details and the allocated seat.
	receipt := &ticket.Receipt{
		TicketId:      ticketID,
		FromLocation:  req.GetFromLocation(),
		ToLocation:    req.GetToLocation(),
		User:          req.GetUser(),
		PricePaid:     req.GetPricePaid(),
		AllocatedSeat: allocatedSeat,
		PurchaseDate:  timestamppb.New(time.Now()),
	}

	// Store the new receipt in our in-memory data structures.
	s.receipts[ticketID] = receipt                           // Store by ticket ID.
	s.occupiedSeats[allocatedSeat.GetSeatNumber()] = receipt // Mark the seat as occupied.

	log.Printf("Ticket purchased successfully. Ticket ID: %s, Allocated Seat: %s (Section %s)",
		ticketID, allocatedSeat.GetSeatNumber(), allocatedSeat.GetSection().String())

	// Return a successful response with the generated receipt.
	return ticket.PurchaseTicketResponse{
		Success: true,
		Message: "Ticket purchased successfully!",
		Receipt: receipt,
	}, nil
}
