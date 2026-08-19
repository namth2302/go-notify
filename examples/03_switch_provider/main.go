package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	notify "github.com/namth/go-notify"
	"github.com/namth/go-notify/message"
	"github.com/namth/go-notify/middleware"
)

func main() {
	ctx := context.Background()

	// Switch provider by simply passing a different configuration.
	// You can load this struct directly from a YAML / JSON file or Viper.
	cfg := notify.Config{
		Provider:  os.Getenv("NOTIFY_PROVIDER"), // e.g. "lark", "gchat", "broadcast", or "fallback"
		Endpoint:  os.Getenv("NOTIFY_WEBHOOK_URL"),
		Secret:    os.Getenv("NOTIFY_SECRET"),
		Timeout:   "5s",
		Retries:   3,
		RateLimit: 5.0,
	}

	notifier, err := notify.NewFromConfig(cfg,
		notify.WithMiddleware(
			middleware.Logging(slog.Default()),
		),
	)
	if err != nil {
		fmt.Printf("Create notifier failed: %v\n", err)
		return
	}

	// Business code doesn't change at all when switching between Google Chat and Lark:
	card := message.NewCard().
		SetStatus(message.StatusInfo).
		SetTitle("📢 Scheduled Maintenance").
		SetSubtitle("Database Cluster Upgrade").
		AddSection(
			message.NewSection().
				AddField("Start Time", "22:00 UTC").
				AddField("Duration", "30 minutes").
				AddField("Impact", "Read-only mode"),
		).
		AddAction(
			message.NewButton("Status Page", "https://status.internal"),
		)

	res, err := notifier.Send(ctx, card.Wrap())
	fmt.Printf("Result: %+v, Err: %v\n", res, err)
}
