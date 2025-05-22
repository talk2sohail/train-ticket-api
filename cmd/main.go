package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

func main() {

	addr := ":9001"

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	client := ticket.NewTrainTicketingServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*200)
	defer cancel()

	respone, err := client.PurchaseTicket(ctx, &ticket.PurchaseTicketRequest{
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
}
