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
		// Initialize necessary fields here
		receipts:      make(map[string]*ticket.Receipt),
		occupiedSeats: make(map[string]*ticket.Receipt),
		sectionCapacities: map[ticket.Seat_Section]int{
			ticket.Seat_SECTION_A: MaxSeatsPerSection,
			ticket.Seat_SECTION_B: MaxSeatsPerSection,
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
		if _, isOccupied := s.occupiedSeats[seatNumber]; !isOccupied {
			return &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: seatNumber,
			}, nil
		}
	}

	// If Section A is full, attempt to find an available seat in Section B.
	for i := 1; i <= s.sectionCapacities[ticket.Seat_SECTION_B]; i++ {
		seatNumber := fmt.Sprintf("B%d", i) // Construct seat string, e.g., "B1", "B2"
		if _, isOccupied := s.occupiedSeats[seatNumber]; !isOccupied {
			return &ticket.Seat{
				Section:    ticket.Seat_SECTION_B,
				SeatNumber: seatNumber,
			}, nil
		}
	}

	return nil, fmt.Errorf("%s", ErrNoAvailableSeats)
}

// PurchaseTicket handles the purchase of a train ticket
func (s *TicketService) PurchaseTicket(ctx context.Context, req *ticket.PurchaseTicketRequest) (ticket.PurchaseTicketResponse, error) {

	// Acquire a lock to protect shared server state (receipts, occupiedSeats) during concurrent access.
	s.mu.Lock()
	defer s.mu.Unlock()

	// find the next available seat using our allocation logic.
	allocatedSeat, err := s.findNextAvailableSeat()
	if err != nil {
		log.Printf("[PurchaseTicket] Failed for user %s: %v", req.GetUser().GetEmail(), err)
		return ticket.PurchaseTicketResponse{
			Success: false,
			Message: err.Error(),
			Receipt: nil,
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
	s.receipts[ticketID] = receipt
	s.occupiedSeats[allocatedSeat.GetSeatNumber()] = receipt

	log.Printf("[PurchaseTicket] Success: TicketID=%s, Seat=%s, Section=%s", ticketID, allocatedSeat.GetSeatNumber(), allocatedSeat.GetSection().String())

	// Return a successful response with the generated receipt.
	return ticket.PurchaseTicketResponse{
		Success: true,
		Message: MsgTicketPurchaseSuccess,
		Receipt: receipt,
	}, nil
}

// GetReceiptDetails retrieves the receipt by ticket ID.
func (s *TicketService) GetReceiptDetails(ctx context.Context, ticketID string) (*ticket.Receipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	receipt, exists := s.receipts[ticketID]
	if !exists {
		err := fmt.Errorf("%s for ticketID %s", ErrReceiptNotFound, ticketID)
		log.Printf("[GetReceiptDetails] %v", err)
		return nil, err
	}
	log.Printf("[GetReceiptDetails] Retrieved receipt for ticketID %s", ticketID)
	return receipt, nil
}

// GetUsersBySection retrieves all users with their seats in a specified section.
func (s *TicketService) GetUsersBySection(ctx context.Context, section ticket.Seat_Section) (ticket.GetUsersBySectionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var users []*ticket.UserSeat
	// Iterate through all receipts to find users in the specified section.
	for _, receipt := range s.receipts {
		if receipt.AllocatedSeat.Section == section {
			users = append(users, &ticket.UserSeat{
				User: receipt.User,
				Seat: receipt.AllocatedSeat,
			})
		}
	}
	log.Printf("[GetUsersBySection] Retrieved %d users in section %s", len(users), section.String())
	return ticket.GetUsersBySectionResponse{
		Success:        true,
		Message:        MsgUsersRetrieved,
		UsersInSection: users,
	}, nil
}

// RemoveUser removes a user identified by their email.
func (s *TicketService) RemoveUser(ctx context.Context, email string) (ticket.RemoveUserResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ticketIdToRemove string
	for id, receipt := range s.receipts {
		if receipt.User.GetEmail() == email {
			ticketIdToRemove = id
			break
		}
	}

	if ticketIdToRemove == "" {
		log.Printf("[RemoveUser] No user found with email: %s", email)
		return ticket.RemoveUserResponse{
			Success: false,
			Message: ErrUserNotFound,
		}, nil
	}

	receipt := s.receipts[ticketIdToRemove]
	delete(s.receipts, ticketIdToRemove)
	delete(s.occupiedSeats, receipt.AllocatedSeat.SeatNumber)
	log.Printf("[RemoveUser] Removed user with email: %s, TicketID: %s", email, ticketIdToRemove)
	return ticket.RemoveUserResponse{
		Success: true,
		Message: MsgUserRemovedSuccess,
	}, nil
}

// ModifyUserSeat updates a user's seat given an existing receipt and the new seat.
func (s *TicketService) ModifyUserSeat(ctx context.Context, receipt *ticket.Receipt, newSeat *ticket.Seat) (ticket.ModifyUserSeatResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure the receipt exists.
	existingUserReceipt, ok := s.receipts[receipt.TicketId]
	if !ok {
		log.Printf("[ModifyUserSeat] Receipt not found for TicketID: %s", receipt.TicketId)
		return ticket.ModifyUserSeatResponse{
			Success: false,
			Message: ErrReceiptNotFound,
		}, nil
	}

	// Check if new seat is occupied by another ticket.
	if occupied, exists := s.occupiedSeats[newSeat.SeatNumber]; exists {
		if occupied.TicketId != receipt.TicketId {
			log.Printf("[ModifyUserSeat] Seat %s is already occupied", newSeat.SeatNumber)
			return ticket.ModifyUserSeatResponse{
				Success: false,
				Message: ErrSeatOccupied,
			}, nil
		}
	}

	// Free the old seat.
	delete(s.occupiedSeats, existingUserReceipt.AllocatedSeat.SeatNumber)
	// Update receipt with the new seat.
	existingUserReceipt.AllocatedSeat = newSeat
	s.occupiedSeats[newSeat.SeatNumber] = existingUserReceipt

	log.Printf("[ModifyUserSeat] Updated seat for TicketID: %s to Seat: %s", receipt.TicketId, newSeat.SeatNumber)
	return ticket.ModifyUserSeatResponse{
		Success:     true,
		Message:     MsgSeatUpdatedSuccess,
		UpdatedSeat: newSeat,
	}, nil
}
