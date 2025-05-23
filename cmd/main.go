package main

import (
	"context"
	"log"
	"time"

	"github.com/talk2sohail/train-ticket-api/client"
	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

func main() {

	addr := ":9001"
	trainTicketClient, err := client.NewTrainTicketClient(addr)
	if err != nil {
		log.Fatalf("could not connect to server: %v", err)
	}
	defer trainTicketClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	respone, err := trainTicketClient.PurchaseTicket(ctx, &ticket.PurchaseTicketRequest{
		FromLocation: "New York",
		ToLocation:   "Los Angeles",
		User: &ticket.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "a@gamil.com",
		},
		PricePaid: 100.0,
	})

	if err != nil {
		log.Fatalf("could not buy the ticket: %v", err)
	}

	log.Printf("Ticket purchased successfully: %s", respone.GetReceipt().GetTicketId())
	log.Printf("Allocated Seat: %s (Section %s)", respone.GetReceipt().GetAllocatedSeat().SeatNumber, respone.GetReceipt().GetAllocatedSeat().GetSection().String())

	// Get receipt details
	receiptDetails, err := trainTicketClient.GetReceiptDetails(ctx, respone.GetReceipt().GetTicketId())
	if err != nil {
		log.Fatalf("could not get receipt details: %v", err)
	}
	// Print receipt details
	log.Printf("Receipt ID: %s", receiptDetails.GetReceipt().GetTicketId())
	log.Printf("From: %s", receiptDetails.GetReceipt().GetFromLocation())
	log.Printf("To: %s", receiptDetails.GetReceipt().GetToLocation())
	log.Printf("User: %s %s (%s)", receiptDetails.GetReceipt().GetUser().GetFirstName(), receiptDetails.GetReceipt().GetUser().GetLastName(), receiptDetails.GetReceipt().GetUser().GetEmail())
	log.Printf("Price Paid: %.2f", receiptDetails.GetReceipt().GetPricePaid())
	log.Printf("Purchase Time: %s", receiptDetails.GetReceipt().GetPurchaseDate().AsTime().Format(time.RFC3339))
	log.Printf("Seat Number: %s", receiptDetails.GetReceipt().GetAllocatedSeat().GetSeatNumber())
	log.Printf("Seat Section: %s", receiptDetails.GetReceipt().GetAllocatedSeat().GetSection().String())

	// Get users by section
	usersBySection, err := trainTicketClient.GetUsersBySection(ctx, ticket.Seat_SECTION_A)
	if err != nil {
		log.Fatalf("could not get users by section: %v", err)
	}
	log.Printf("Users in Section A: %v", usersBySection.GetUsersInSection())

	// Get users by section
	usersBySection, err = trainTicketClient.GetUsersBySection(ctx, ticket.Seat_SECTION_B)
	if err != nil {
		log.Fatalf("could not get users by section: %v", err)
	}
	log.Printf("Users in Section B: %v", usersBySection.GetUsersInSection())

	// Modify user seat
	newSeat := &ticket.Seat{
		SeatNumber: "B2",
		Section:    ticket.Seat_SECTION_A,
	}
	modifiedSeat, err := trainTicketClient.ModifyUserSeat(ctx, receiptDetails.GetReceipt().GetTicketId(), newSeat)
	if err != nil {
		log.Fatalf("could not modify user seat: %v", err)
	}
	log.Printf("User seat modified successfully: %s", modifiedSeat.GetMessage())
	log.Printf("New Seat Number: %s", modifiedSeat.GetUpdatedSeat().GetSeatNumber())

	// Remove user
	email := receiptDetails.GetReceipt().GetUser().GetEmail()
	removedUser, err := trainTicketClient.RemoveUser(ctx, email)
	if err != nil {
		log.Fatalf("could not remove user: %v", err)
	}

	log.Printf("User removed successfully: %s", removedUser.GetMessage())

}
