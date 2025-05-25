package service

import (
	"context"
	"fmt"
	"sync"
	"testing"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

func TestUnit_FindNextAvailableSeat(t *testing.T) {
	s := NewTicketService()

	t.Run("Section A available", func(t *testing.T) {
		seat, err := s.findNextAvailableSeat()
		if err != nil {
			t.Fatalf("expected seat, got error: %v", err)
		}
		if seat.Section != ticket.Seat_SECTION_A || seat.SeatNumber != "A1" {
			t.Errorf("expected seat A1 in Section A, got %s in Section %s", seat.SeatNumber, seat.Section.String())
		}
	})

	t.Run("Section A full, Section B available", func(t *testing.T) {
		for i := 1; i <= s.sectionCapacities[ticket.Seat_SECTION_A]; i++ {
			seatNum := fmt.Sprintf("A%d", i)
			s.occupiedSeats[seatNum] = &ticket.Receipt{}
		}
		seat, err := s.findNextAvailableSeat()
		if err != nil {
			t.Fatalf("expected seat in Section B, got error: %v", err)
		}
		if seat.Section != ticket.Seat_SECTION_B || seat.SeatNumber != "B1" {
			t.Errorf("expected seat B1 in Section B, got %s in Section %s", seat.SeatNumber, seat.Section.String())
		}
	})

	t.Run("No available seats", func(t *testing.T) {
		for i := 1; i <= s.sectionCapacities[ticket.Seat_SECTION_B]; i++ {
			seatNum := fmt.Sprintf("B%d", i)
			s.occupiedSeats[seatNum] = &ticket.Receipt{}
		}
		seat, err := s.findNextAvailableSeat()
		if err == nil {
			t.Fatalf("expected error %s, got seat: %v", ErrNoAvailableSeats, seat)
		}
		if err.Error() != ErrNoAvailableSeats {
			t.Errorf("expected error %s, got %s", ErrNoAvailableSeats, err.Error())
		}
	})
}

func TestUnit_PurchaseTicket(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful purchase", func(t *testing.T) {
		s := NewTicketService()
		req := &ticket.PurchaseTicketRequest{
			FromLocation: "London",
			ToLocation:   "Paris",
			User: &ticket.User{
				FirstName: "Test",
				LastName:  "User",
				Email:     "test@example.com",
			},
			PricePaid: 50.0,
		}
		res, err := s.PurchaseTicket(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !res.Success {
			t.Fatalf("expected success, got failure")
		}
		if res.Receipt == nil {
			t.Fatalf("expected a receipt, got nil")
		}
		if res.Receipt.AllocatedSeat.SeatNumber != "A1" {
			t.Errorf("expected allocated seat 'A1', got: %s", res.Receipt.AllocatedSeat.SeatNumber)
		}
	})

	t.Run("Purchase fails when no seats available", func(t *testing.T) {
		s := NewTicketService()
		for _, section := range []ticket.Seat_Section{ticket.Seat_SECTION_A, ticket.Seat_SECTION_B} {
			for i := 1; i <= s.sectionCapacities[section]; i++ {
				seatID := ""
				if section == ticket.Seat_SECTION_A {
					seatID = fmt.Sprintf("A%d", i)
				} else {
					seatID = fmt.Sprintf("B%d", i)
				}
				s.occupiedSeats[seatID] = &ticket.Receipt{}
			}
		}

		req := &ticket.PurchaseTicketRequest{
			FromLocation: "London",
			ToLocation:   "Paris",
			User: &ticket.User{
				FirstName: "Test",
				LastName:  "User",
				Email:     "test2@example.com",
			},
			PricePaid: 50.0,
		}
		res, err := s.PurchaseTicket(ctx, req)
		if err != nil {
			t.Fatalf("expected no error returned, got: %v", err)
		}
		if res.Success {
			t.Fatalf("expected purchase failure, but got success")
		}
		if res.Receipt != nil {
			t.Errorf("expected no receipt on failure, got one")
		}
		if res.Message != ErrNoAvailableSeats {
			t.Errorf("expected message %q, got: %q", ErrNoAvailableSeats, res.Message)
		}
	})

	t.Run("Multiple sequential purchases", func(t *testing.T) {
		s := NewTicketService()
		totalPurchases := 3
		expectedSeats := []string{"A1", "A2", "A3"}
		for i := 0; i < totalPurchases; i++ {
			req := &ticket.PurchaseTicketRequest{
				FromLocation: "CityX",
				ToLocation:   "CityY",
				User: &ticket.User{
					FirstName: fmt.Sprintf("User%d", i),
					LastName:  "Test",
					Email:     fmt.Sprintf("user%d@example.com", i),
				},
				PricePaid: 30.0,
			}
			res, err := s.PurchaseTicket(ctx, req)
			if err != nil {
				t.Fatalf("unexpected error on purchase %d: %v", i, err)
			}
			if !res.Success {
				t.Fatalf("expected success on purchase %d, got failure: %s", i, res.Message)
			}
			if res.Receipt.AllocatedSeat.SeatNumber != expectedSeats[i] {
				t.Errorf("expected seat %s, got %s", expectedSeats[i], res.Receipt.AllocatedSeat.SeatNumber)
			}
		}
	})
}

func TestUnit_PurchaseTicketConcurrent(t *testing.T) {
	ctx := context.Background()
	s := NewTicketService()

	totalSeats := s.sectionCapacities[ticket.Seat_SECTION_A] + s.sectionCapacities[ticket.Seat_SECTION_B]

	type result struct {
		response ticket.PurchaseTicketResponse
		err      error
	}
	var wg sync.WaitGroup

	numAttempts := totalSeats + 2
	emailPrefix := "user_%d@example.com"
	var resChan = make(chan result, numAttempts)
	wg.Add(numAttempts)
	for i := 0; i < numAttempts; i++ {
		go func(i int) {
			defer wg.Done()
			req := &ticket.PurchaseTicketRequest{
				FromLocation: "CityX",
				ToLocation:   "CityY",
				User: &ticket.User{
					FirstName: "Concurrent",
					LastName:  "User",
					Email:     fmt.Sprintf(emailPrefix, i),
				},
				PricePaid: 25.0,
			}
			res, err := s.PurchaseTicket(ctx, req)
			resChan <- result{response: res, err: err}
		}(i)
	}
	wg.Wait()
	close(resChan)

	successMap := make(map[string]bool)
	var successCount int
	for res := range resChan {
		if res.err != nil {
			t.Errorf("Unexpected error: %v", res.err)
			continue
		}
		if res.response.Success {
			successCount++
			seat := res.response.Receipt.AllocatedSeat.SeatNumber
			if successMap[seat] {
				t.Errorf("Duplicate allocation for seat %s", seat)
			}
			successMap[seat] = true
		} else {
			if res.response.Message != ErrNoAvailableSeats {
				t.Errorf("Expected error message %q, got %q", ErrNoAvailableSeats, res.response.Message)
			}
		}
	}

	if successCount != totalSeats {
		t.Errorf("Expected %d successful purchases, got %d", totalSeats, successCount)
	}
}

func TestUnit_GetReceiptDetails(t *testing.T) {
	ctx := context.Background()
	s := NewTicketService()

	t.Run("Receipt exists", func(t *testing.T) {
		ticketID := "test-ticket-1"
		expectedReceipt := &ticket.Receipt{
			TicketId:     ticketID,
			FromLocation: "CityA",
			ToLocation:   "CityB",
			PricePaid:    100.0,
		}
		s.receipts[ticketID] = expectedReceipt

		receipt, err := s.GetReceiptDetails(ctx, ticketID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if receipt.TicketId != expectedReceipt.TicketId {
			t.Errorf("expected TicketId %s, got %s", expectedReceipt.TicketId, receipt.TicketId)
		}
		if receipt.FromLocation != expectedReceipt.FromLocation {
			t.Errorf("expected FromLocation %s, got %s", expectedReceipt.FromLocation, receipt.FromLocation)
		}
	})

	t.Run("Receipt does not exist", func(t *testing.T) {
		unknownTicketID := "non-existent-ticket"
		_, err := s.GetReceiptDetails(ctx, unknownTicketID)
		if err == nil {
			t.Fatalf("expected error for ticketID %s, got nil", unknownTicketID)
		}
		expectedError := fmt.Sprintf("%s for ticketID %s", ErrReceiptNotFound, unknownTicketID)
		if err.Error() != expectedError {
			t.Errorf("expected error %q, got %q", expectedError, err.Error())
		}
	})

	t.Run("Empty TicketID returns error", func(t *testing.T) {
		emptyID := ""
		_, err := s.GetReceiptDetails(ctx, emptyID)
		if err == nil {
			t.Fatalf("expected error for empty ticketID, got nil")
		}
		expected := fmt.Sprintf("%s for ticketID %s", ErrReceiptNotFound, emptyID)
		if err.Error() != expected {
			t.Errorf("expected error %q, got %q", expected, err.Error())
		}
	})

	t.Run("Concurrent retrieval of a receipt", func(t *testing.T) {
		ticketID := "concurrent-ticket"
		expectedReceipt := &ticket.Receipt{
			TicketId:     ticketID,
			FromLocation: "CityX",
			ToLocation:   "CityY",
			PricePaid:    75.0,
		}
		s.receipts[ticketID] = expectedReceipt

		numGoroutines := 20
		var wg sync.WaitGroup
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				receipt, err := s.GetReceiptDetails(ctx, ticketID)
				if err != nil {
					errChan <- err
					return
				}
				if receipt.TicketId != expectedReceipt.TicketId {
					errChan <- err
				}
			}()
		}
		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				t.Errorf("concurrent retrieval error: %v", err)
			}
		}
	})
}

func TestUnit_GetUsersBySection(t *testing.T) {
	ctx := context.Background()
	s := NewTicketService()

	t.Run("Empty section returns no users", func(t *testing.T) {
		resp, err := s.GetUsersBySection(ctx, ticket.Seat_SECTION_A)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(resp.UsersInSection) != 0 {
			t.Errorf("expected 0 users, got %d", len(resp.UsersInSection))
		}
	})

	t.Run("Returns users in specified section", func(t *testing.T) {
		r1 := &ticket.Receipt{
			TicketId: "ticket1",
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: "A1",
			},
			User: &ticket.User{Email: "user1@example.com"},
		}
		r2 := &ticket.Receipt{
			TicketId: "ticket2",
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: "A2",
			},
			User: &ticket.User{Email: "user2@example.com"},
		}
		r3 := &ticket.Receipt{
			TicketId: "ticket3",
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_B,
				SeatNumber: "B1",
			},
			User: &ticket.User{Email: "user3@example.com"},
		}

		s.receipts[r1.TicketId] = r1
		s.receipts[r2.TicketId] = r2
		s.receipts[r3.TicketId] = r3

		s.occupiedSeats[r1.AllocatedSeat.SeatNumber] = r1
		s.occupiedSeats[r2.AllocatedSeat.SeatNumber] = r2
		s.occupiedSeats[r3.AllocatedSeat.SeatNumber] = r3

		resp, err := s.GetUsersBySection(ctx, ticket.Seat_SECTION_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.UsersInSection) != 2 {
			t.Errorf("expected 2 users in Section A, got %d", len(resp.UsersInSection))
		}
		emails := map[string]bool{
			"user1@example.com": false,
			"user2@example.com": false,
		}
		for _, us := range resp.UsersInSection {
			if _, ok := emails[us.User.GetEmail()]; ok {
				emails[us.User.GetEmail()] = true
			}
		}
		for email, found := range emails {
			if !found {
				t.Errorf("expected email %s in results", email)
			}
		}
	})
}

func TestUnit_RemoveUser(t *testing.T) {
	ctx := context.Background()
	s := NewTicketService()

	t.Run("Removes existing user", func(t *testing.T) {
		receipt := &ticket.Receipt{
			TicketId: "ticket1",
			User: &ticket.User{
				FirstName: "Test",
				LastName:  "User",
				Email:     "user@example.com",
			},
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: "A1",
			},
		}
		s.receipts["ticket1"] = receipt
		s.occupiedSeats["A1"] = receipt

		resp, err := s.RemoveUser(ctx, "user@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !resp.Success {
			t.Errorf("expected removal success, got failure with message: %s", resp.Message)
		}

		if _, exists := s.receipts["ticket1"]; exists {
			t.Errorf("expected receipt to be removed")
		}
		if _, exists := s.occupiedSeats["A1"]; exists {
			t.Errorf("expected seat to be unoccupied")
		}
	})

	t.Run("User not found", func(t *testing.T) {

		resp, err := s.RemoveUser(ctx, "nonexistent@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Success {
			t.Errorf("expected failure when removing nonexistent user")
		}
		if resp.Message != ErrUserNotFound {
			t.Errorf("expected error message %q, got: %q", ErrUserNotFound, resp.Message)
		}
	})
}

func TestUnit_ModifyUserSeat(t *testing.T) {
	ctx := context.Background()
	s := NewTicketService()

	t.Run("Successful seat modification", func(t *testing.T) {
		receipt := &ticket.Receipt{
			TicketId: "ticket1",
			User: &ticket.User{
				Email: "user@example.com",
			},
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: "A1",
			},
		}
		s.receipts[receipt.TicketId] = receipt
		s.occupiedSeats["A1"] = receipt

		newSeat := &ticket.Seat{
			Section:    ticket.Seat_SECTION_A,
			SeatNumber: "A2",
		}
		resp, err := s.ModifyUserSeat(ctx, receipt, newSeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Success {
			t.Fatalf("expected success, got failure with message: %s", resp.Message)
		}
		if receipt.AllocatedSeat.SeatNumber != "A2" {
			t.Errorf("expected seat to be updated to A2, got %s", receipt.AllocatedSeat.SeatNumber)
		}
		if _, exists := s.occupiedSeats["A1"]; exists {
			t.Errorf("expected seat A1 to be freed")
		}
	})

	t.Run("Fail modification due to seat already occupied", func(t *testing.T) {
		receipt1 := &ticket.Receipt{
			TicketId: "ticket2",
			User: &ticket.User{
				Email: "first@example.com",
			},
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_B,
				SeatNumber: "B1",
			},
		}
		receipt2 := &ticket.Receipt{
			TicketId: "ticket3",
			User: &ticket.User{
				Email: "second@example.com",
			},
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_B,
				SeatNumber: "B2",
			},
		}
		s.receipts[receipt1.TicketId] = receipt1
		s.occupiedSeats["B1"] = receipt1

		s.receipts[receipt2.TicketId] = receipt2
		s.occupiedSeats["B2"] = receipt2

		newSeat := &ticket.Seat{
			Section:    ticket.Seat_SECTION_B,
			SeatNumber: "B2",
		}
		resp, err := s.ModifyUserSeat(ctx, receipt1, newSeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Success {
			t.Fatalf("expected failure due to seat being occupied")
		}
		if resp.Message != ErrSeatOccupied {
			t.Errorf("expected error message %q, got %q", ErrSeatOccupied, resp.Message)
		}
	})

	t.Run("Fail modification due to non-existent receipt", func(t *testing.T) {
		receipt := &ticket.Receipt{
			TicketId: "nonexistent_ticket",
			User: &ticket.User{
				Email: "nonexistent@example.com",
			},
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: "A3",
			},
		}
		newSeat := &ticket.Seat{
			Section:    ticket.Seat_SECTION_A,
			SeatNumber: "A4",
		}
		resp, err := s.ModifyUserSeat(ctx, receipt, newSeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Success {
			t.Fatalf("expected failure due to non-existent receipt")
		}
		if resp.Message != ErrReceiptNotFound {
			t.Errorf("expected error message %q, got %q", ErrReceiptNotFound, resp.Message)
		}
	})

	t.Run("Modify seat to same seat returns success", func(t *testing.T) {
		receipt := &ticket.Receipt{
			TicketId: "ticket_same",
			User: &ticket.User{
				Email: "same@example.com",
			},
			AllocatedSeat: &ticket.Seat{
				Section:    ticket.Seat_SECTION_A,
				SeatNumber: "A1",
			},
		}
		s.receipts[receipt.TicketId] = receipt
		s.occupiedSeats["A1"] = receipt

		sameSeat := &ticket.Seat{
			Section:    ticket.Seat_SECTION_A,
			SeatNumber: "A1",
		}
		resp, err := s.ModifyUserSeat(ctx, receipt, sameSeat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.Success {
			t.Fatalf("expected success when modifying to the same seat, got failure with message: %s", resp.Message)
		}
		if receipt.AllocatedSeat.SeatNumber != "A1" {
			t.Errorf("expected seat to remain A1, got %s", receipt.AllocatedSeat.SeatNumber)
		}
	})
}
