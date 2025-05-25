package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	ticket "github.com/talk2sohail/train-ticket-api/internal/common/genproto/ticket"
	"github.com/talk2sohail/train-ticket-api/internal/ticket/handler"
	"github.com/talk2sohail/train-ticket-api/internal/ticket/service"
	"github.com/talk2sohail/train-ticket-api/mock"
)

func TestUnit_HandlerPurchaseTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	validReq := &ticket.PurchaseTicketRequest{
		FromLocation: "Station A",
		ToLocation:   "Station B",
		User: &ticket.User{
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice.smith@example.com",
		},
		PricePaid: 50.0,
	}

	t.Run("invalid request", func(t *testing.T) {
		emptyReq := &ticket.PurchaseTicketRequest{}
		mockSvc := mock.NewMockTicketService(ctrl)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.PurchaseTicket(ctx, emptyReq)
		if err == nil {
			t.Errorf("expected error for invalid request, got nil")
		}
	})

	t.Run("service error", func(t *testing.T) {
		expectedErr := errors.New("service failure")
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().PurchaseTicket(ctx, validReq).Return(ticket.PurchaseTicketResponse{}, expectedErr)

		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.PurchaseTicket(ctx, validReq)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if resp != nil {
			t.Errorf("expected nil response on error, got %v", resp)
		}
	})

	t.Run("successful purchase", func(t *testing.T) {
		expectedResp := ticket.PurchaseTicketResponse{
			Message: service.MsgTicketPurchaseSuccess,
			Success: true,
			Receipt: &ticket.Receipt{
				TicketId: "12345",
			},
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().PurchaseTicket(ctx, validReq).Return(expectedResp, nil)

		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.PurchaseTicket(ctx, validReq)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Errorf("expected a non-nil response")
		}

		if resp.GetMessage() != expectedResp.GetMessage() {
			t.Errorf("expected message %s, got %s", expectedResp.Message, resp.Message)
		}
		if resp.GetSuccess() != expectedResp.GetSuccess() {
			t.Errorf("expected success %v, got %v", expectedResp.Success, resp.Success)
		}
		if resp.GetReceipt().GetTicketId() != expectedResp.Receipt.GetTicketId() {
			t.Errorf("expected ticket ID %s, got %s", expectedResp.Receipt.GetTicketId(), resp.GetReceipt().GetTicketId())
		}
	})

}

func TestUnit_HandlerGetReceiptDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	validReq := &ticket.GetReceiptDetailsRequest{
		Identifier: &ticket.GetReceiptDetailsRequest_TicketId{TicketId: "ticket-123"},
	}

	t.Run("service error", func(t *testing.T) {
		expectedErr := errors.New("receipt retrieval failed")
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetReceiptDetails(ctx, "ticket-123").
			Return(nil, expectedErr)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.GetReceiptDetails(ctx, validReq)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if resp != nil {
			t.Errorf("expected nil response on error, got %v", resp)
		}
	})

	t.Run("successful retrieval", func(t *testing.T) {
		expectedReceipt := &ticket.Receipt{
			TicketId: "ticket-123",
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetReceiptDetails(ctx, "ticket-123").
			Return(expectedReceipt, nil)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.GetReceiptDetails(ctx, validReq)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp == nil || resp.GetReceipt().GetTicketId() != expectedReceipt.GetTicketId() {
			t.Errorf("expected receipt ticketId %s, got %v", expectedReceipt.TicketId, resp)
		}
	})

}

func TestUnit_HandlerGetUsersBySection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("invalid section", func(t *testing.T) {
		req := &ticket.GetUsersBySectionRequest{
			Section: ticket.Seat_SECTION_UNKNOWN,
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.GetUsersBySection(ctx, req)
		if err == nil {
			t.Errorf("expected error for invalid section, got nil")
		}
	})

	t.Run("service error", func(t *testing.T) {
		req := &ticket.GetUsersBySectionRequest{
			Section: ticket.Seat_SECTION_A,
		}
		expectedErr := errors.New("service failure")
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetUsersBySection(ctx, ticket.Seat_SECTION_A).
			Return(ticket.GetUsersBySectionResponse{}, expectedErr)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.GetUsersBySection(ctx, req)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if resp != nil {
			t.Errorf("expected nil response on error, got %v", resp)
		}
	})

	t.Run("successful retrieval", func(t *testing.T) {
		req := &ticket.GetUsersBySectionRequest{
			Section: ticket.Seat_SECTION_A,
		}
		expectedResp := ticket.GetUsersBySectionResponse{}
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetUsersBySection(ctx, ticket.Seat_SECTION_A).
			Return(expectedResp, nil)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.GetUsersBySection(ctx, req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Errorf("expected a non-nil response")
		}
	})
}

func TestUnit_HandlerRemoveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("missing email", func(t *testing.T) {
		req := &ticket.RemoveUserRequest{
			Identifier: &ticket.RemoveUserRequest_Email{Email: ""},
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.RemoveUser(ctx, req)
		if err == nil {
			t.Errorf("expected error for missing email, got nil")
		}
	})

	t.Run("service error", func(t *testing.T) {
		req := &ticket.RemoveUserRequest{
			Identifier: &ticket.RemoveUserRequest_Email{Email: "user@example.com"},
		}
		expectedErr := errors.New("service removal error")
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			RemoveUser(ctx, "user@example.com").
			Return(ticket.RemoveUserResponse{}, expectedErr)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.RemoveUser(ctx, req)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if resp != nil {
			t.Errorf("expected nil response on error, got %v", resp)
		}
	})

	t.Run("successful removal", func(t *testing.T) {
		req := &ticket.RemoveUserRequest{
			Identifier: &ticket.RemoveUserRequest_Email{Email: "user@example.com"},
		}
		expectedResp := ticket.RemoveUserResponse{}
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			RemoveUser(ctx, "user@example.com").
			Return(expectedResp, nil)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.RemoveUser(ctx, req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Errorf("expected a non-nil response")
		}
	})
}

func TestUnit_HandlerModifyUserSeat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("missing ticketId", func(t *testing.T) {
		req := &ticket.ModifyUserSeatRequest{
			Identifier: &ticket.ModifyUserSeatRequest_TicketId{TicketId: ""},
			NewSeat:    &ticket.Seat{SeatNumber: "A1"},
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.ModifyUserSeat(ctx, req)
		if err == nil {
			t.Errorf("expected error for missing ticketId, got nil")
		}
	})

	t.Run("missing new seat", func(t *testing.T) {
		req := &ticket.ModifyUserSeatRequest{
			Identifier: &ticket.ModifyUserSeatRequest_TicketId{TicketId: "ticket-123"},
			NewSeat:    nil,
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.ModifyUserSeat(ctx, req)
		if err == nil {
			t.Errorf("expected error for missing new seat, got nil")
		}
	})

	t.Run("service error on receipt retrieval", func(t *testing.T) {
		req := &ticket.ModifyUserSeatRequest{
			Identifier: &ticket.ModifyUserSeatRequest_TicketId{TicketId: "ticket-123"},
			NewSeat:    &ticket.Seat{SeatNumber: "B2"},
		}
		expectedErr := errors.New("failed to retrieve receipt")
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetReceiptDetails(ctx, "ticket-123").
			Return(nil, expectedErr)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.ModifyUserSeat(ctx, req)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("service error on seat modification", func(t *testing.T) {
		req := &ticket.ModifyUserSeatRequest{
			Identifier: &ticket.ModifyUserSeatRequest_TicketId{TicketId: "ticket-123"},
			NewSeat:    &ticket.Seat{SeatNumber: "B2"},
		}
		expectedReceipt := &ticket.Receipt{TicketId: "ticket-123"}
		expectedErr := errors.New("seat modification failed")
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetReceiptDetails(ctx, "ticket-123").
			Return(expectedReceipt, nil)
		mockSvc.EXPECT().
			ModifyUserSeat(ctx, expectedReceipt, req.GetNewSeat()).
			Return(ticket.ModifyUserSeatResponse{}, expectedErr)
		h := handler.NewTicketGrpcHandler(mockSvc)
		_, err := h.ModifyUserSeat(ctx, req)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("successful modification", func(t *testing.T) {
		req := &ticket.ModifyUserSeatRequest{
			Identifier: &ticket.ModifyUserSeatRequest_TicketId{TicketId: "ticket-123"},
			NewSeat:    &ticket.Seat{SeatNumber: "B2"},
		}
		expectedReceipt := &ticket.Receipt{TicketId: "ticket-123"}
		expectedResp := ticket.ModifyUserSeatResponse{
			Message: "seat updated successfully",
		}
		mockSvc := mock.NewMockTicketService(ctrl)
		mockSvc.EXPECT().
			GetReceiptDetails(ctx, "ticket-123").
			Return(expectedReceipt, nil)
		mockSvc.EXPECT().
			ModifyUserSeat(ctx, expectedReceipt, req.GetNewSeat()).
			Return(expectedResp, nil)
		h := handler.NewTicketGrpcHandler(mockSvc)
		resp, err := h.ModifyUserSeat(ctx, req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Errorf("expected a non-nil response")
		}
		if resp.GetMessage() != expectedResp.Message {
			t.Errorf("expected message %s, got %s", expectedResp.Message, resp.GetMessage())
		}
	})
}
