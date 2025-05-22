package util

import (
	"fmt"
	"log"

	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
)

func ValidatePurchseRequestObject(r *ticket.PurchaseTicketRequest) error {
	if r.GetFromLocation() == "" {
		log.Printf("FromLocation is required")
		return fmt.Errorf("FromLocation is required")
	}
	if r.GetToLocation() == "" {
		log.Printf("ToLocation is required")
		return fmt.Errorf("ToLocation is required")
	}
	if r.GetUser() == nil {
		log.Printf("User is required")
		return fmt.Errorf("User is required")
	}
	if r.GetPricePaid() <= 0 {
		log.Printf("PricePaid must be greater than zero")
		return fmt.Errorf("PricePaid must be greater than zero")
	}
	return nil
}

func ValidateSection(r *ticket.GetUsersBySectionRequest) error {
	if r.GetSection().String() == "" {
		log.Printf("Section is required")
		return fmt.Errorf("Section is required")
	}
	return nil
}
