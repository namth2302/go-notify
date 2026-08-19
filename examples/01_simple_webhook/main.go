package main

import (
	"context"
	"fmt"
	"os"

	notify "github.com/namth2302/go-notify"
	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/provider/gchat"
	"github.com/namth2302/go-notify/provider/lark"
)

func main() {
	ctx := context.Background()

	// 1. Send Rich Card to Google Chat
	gchatSender, err := gchat.New(
		gchat.WithWebhookURL(os.Getenv("GCHAT_WEBHOOK_URL")),
	)
	if err == nil {
		card := message.NewCard().
			SetStatus(message.StatusWarning).
			SetTitle("🚨 High CPU Usage").
			SetSubtitle("Service: payment-gateway | Env: prod").
			AddSection(
				message.NewSection().
					SetHeader("Resource Details").
					AddField("CPU Usage", "94.2%").
					AddField("Memory", "6.8GB / 8GB").
					AddButton(message.NewButton("View Grafana", "https://grafana.internal/d/123")),
			).
			AddAction(
				message.NewButton("Acknowledge", "https://ops.internal/ack/123").AsPrimary(),
			)

		res, err := gchatSender.Send(ctx, card.Wrap())
		fmt.Printf("GChat result: res=%+v err=%v\n", res, err)
	}

	// 2. Send Rich Card to Lark (Feishu) with HMAC Secret
	larkSender, err := lark.New(
		lark.WithWebhookURL(os.Getenv("LARK_WEBHOOK_URL")),
		lark.WithSecret(os.Getenv("LARK_SECRET")),
	)
	if err == nil {
		n := notify.New(larkSender)
		res, err := n.Send(ctx, message.NewText("Simple plain text alert to Lark"))
		fmt.Printf("Lark result: res=%+v err=%v\n", res, err)
	}
}
