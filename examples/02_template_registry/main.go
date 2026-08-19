package main

import (
	"context"
	"fmt"
	"os"

	notify "github.com/namth2302/go-notify"
	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/template"
)

// PaymentEvent is a typed DTO
type PaymentEvent struct {
	OrderID       string
	CustomerEmail string
	Amount        string
	Method        string
	Success       bool
}

func init() {
	// Pattern 1: Declarative Card Template using standard Go {{ .Var }}
	_ = notify.RegisterTemplate("deploy_completed", template.NewCardTemplate().
		SetStatus(message.StatusSuccess).
		SetTitle("🚀 Deploy Thành Công: {{ .ServiceName }}").
		SetSubtitle("Version: {{ .Version }} | Env: {{ .Env }}").
		AddSection(
			template.NewSectionTemplate().
				AddField("Commit", "{{ .CommitSHA }}").
				AddField("Author", "{{ .Author }}"),
		).
		AddAction(
			template.NewButtonTemplate("Xem Logs", "https://k8s.internal/deploy/{{ .ServiceName }}").AsPrimary(),
		),
	)

	// Pattern 2: Type-safe Functional Template using Generics
	_ = notify.RegisterFunc("payment_event", func(event PaymentEvent) *message.Card {
		status := message.StatusSuccess
		title := fmt.Sprintf("💳 Thanh toán thành công #%s", event.OrderID)
		if !event.Success {
			status = message.StatusDanger
			title = fmt.Sprintf("❌ Thanh toán thất bại #%s", event.OrderID)
		}

		return message.NewCard().
			SetStatus(status).
			SetTitle(title).
			AddSection(
				message.NewSection().
					AddField("Khách hàng", event.CustomerEmail).
					AddField("Số tiền", event.Amount).
					AddField("Hình thức", event.Method),
			)
	})
}

func main() {
	ctx := context.Background()

	// Initialize Notifier from Environment
	sender, err := notify.NewFromEnv()
	if err != nil {
		fmt.Printf("Init sender failed: %v\n", err)
		return
	}

	// 1. Fast send using Declarative template with Map
	_, _ = sender.SendTemplate(ctx, "deploy_completed", map[string]string{
		"ServiceName": "order-service",
		"Version":     "v2.10.4",
		"Env":         "Production",
		"CommitSHA":   "a9f4e21",
		"Author":      os.Getenv("USER"),
	})

	// 2. Fast send using Type-Safe template with Struct
	_, _ = sender.SendTemplate(ctx, "payment_event", PaymentEvent{
		OrderID:       "ORD-88741",
		CustomerEmail: "customer@example.com",
		Amount:        "1,250,000 VND",
		Method:        "VNPay QR",
		Success:       true,
	})
}
