package middleware

import (
	"context"
	"log/slog"

	"github.com/namth/go-notify/message"
)

// Logging returns a middleware that logs notification requests with slog.
func Logging(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, msg message.Message) (*message.Result, error) {
			msgType := "unknown"
			if msg != nil {
				msgType = string(msg.Type())
			}

			logger.DebugContext(ctx, "sending notification",
				slog.String("message_type", msgType),
			)

			res, err := next(ctx, msg)

			if err != nil {
				provider := "unknown"
				if res != nil && res.Provider != "" {
					provider = res.Provider
				}
				logger.ErrorContext(ctx, "notification send failed",
					slog.String("provider", provider),
					slog.String("message_type", msgType),
					slog.Any("error", err),
				)
				return res, err
			}

			logger.InfoContext(ctx, "notification sent successfully",
				slog.String("provider", res.Provider),
				slog.Int("status_code", res.StatusCode),
				slog.Duration("duration", res.Duration),
			)

			return res, nil
		}
	}
}
