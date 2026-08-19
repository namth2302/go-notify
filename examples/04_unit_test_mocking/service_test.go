package main_test

import (
	"context"
	"fmt"
	"testing"

	notify "github.com/namth/go-notify"
	"github.com/namth/go-notify/message"
	"github.com/namth/go-notify/testutil"
)

// OrderService is an example consumer business service.
type OrderService struct {
	notifier notify.Sender // Inject interface
}

func NewOrderService(notifier notify.Sender) *OrderService {
	return &OrderService{notifier: notifier}
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID string, reason string) error {
	// ... Perform database cancellation ...

	// Send notification
	card := message.NewCard().
		SetStatus(message.StatusDanger).
		SetTitle(fmt.Sprintf("❌ Đơn hàng đã hủy: %s", orderID)).
		AddSection(
			message.NewSection().
				AddField("Mã đơn", orderID).
				AddField("Lý do", reason),
		)

	_, err := s.notifier.Send(ctx, card.Wrap())
	return err
}

func TestOrderService_CancelOrder(t *testing.T) {
	// Arrange
	mockNotifier := testutil.NewMockSender()
	recorder := testutil.NewRecorder(mockNotifier)
	service := NewOrderService(mockNotifier)

	// Act
	err := service.CancelOrder(context.Background(), "ORD-12345", "Khách hàng đổi ý")
	if err != nil {
		t.Fatalf("cancel order failed: %v", err)
	}

	// Assert
	if mockNotifier.SentCount() != 1 {
		t.Fatalf("expected 1 notification sent, got %d", mockNotifier.SentCount())
	}

	if !recorder.HasTitle("❌ Đơn hàng đã hủy: ORD-12345") {
		t.Errorf("expected card title to match cancelled order")
	}

	if !recorder.HasStatus(message.StatusDanger) {
		t.Errorf("expected alert status danger")
	}

	if !recorder.ContainsText("Khách hàng đổi ý") {
		t.Errorf("expected message to contain cancellation reason")
	}
}
