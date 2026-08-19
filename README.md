# go-notify

[![Go Reference](https://pkg.go.dev/badge/github.com/namth2302/go-notify.svg)](https://pkg.go.dev/github.com/namth2302/go-notify)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/Dependencies-Zero-brightgreen.svg)](go.mod)

A simple, fast, and lightweight Go library to send notifications to **Google Chat** and **Lark (Feishu)**.

---

## ✨ Features

- **Zero Dependencies**: Uses only standard Go packages. Keeps your project small and fast.
- **One Card for All**: Write a message card once. It works on both Google Chat and Lark automatically.
- **Fast Switch**: Change between Google Chat and Lark with just one line of configuration.
- **Template Registry**: Save your message templates and send them quickly with your data.
- **Auto Retry & Rate Limit**: Retries automatically when network fails and stops sending too fast.
- **Easy to Test**: Includes mock tools to test your app without sending real messages.

---

## 🚀 Installation

```bash
go get github.com/namth2302/go-notify
```

---

## 📖 Quickstart

### Send a Card Message

```go
package main

import (
    "context"
    "os"

    notify "github.com/namth2302/go-notify"
    "github.com/namth2302/go-notify/message"
)

func main() {
    ctx := context.Background()

    // 1. Create a sender from environment variables:
    // Set NOTIFY_PROVIDER=gchat (or lark) and NOTIFY_ENDPOINT=your_webhook_url
    sender, err := notify.NewFromEnv()
    if err != nil {
        panic(err)
    }

    // 2. Build your card message
    card := message.NewCard().
        SetStatus(message.StatusDanger). // Shows Red color automatically
        SetTitle("🚨 High CPU Alert").
        SetSubtitle("Service: payment-api | Env: Production").
        AddSection(
            message.NewSection().
                AddField("CPU Usage", "94%").
                AddField("Threshold", "85%").
                AddField("Host", "server-01"),
        ).
        AddAction(
            message.NewButton("Open Dashboard", "https://grafana.internal/d/123"),
            message.NewButton("Acknowledge", "https://ops.internal/ack").AsPrimary(),
        )

    // 3. Send the message
    res, err := sender.Send(ctx, card.Wrap())
    if err != nil {
        panic(err)
    }
    _ = res
}
```

---

## 📑 Message Templates (Send Fast)

You can register a template once when your app starts, and use it anywhere with your data.

### Option 1: Text Template with `{{ .Variables }}`

```go
// 1. Register template at startup
notify.RegisterTemplate("deploy_done", template.NewCardTemplate().
    SetStatus(message.StatusSuccess).
    SetTitle("🚀 Deploy Success: {{ .Service }}").
    AddSection(
        template.NewSectionTemplate().
            AddField("Version", "{{ .Version }}").
            AddField("Environment", "{{ .Env }}"),
    ),
)

// 2. Send fast by passing your data (map or struct)
_ = sender.SendTemplate(ctx, "deploy_done", map[string]string{
    "Service": "order-api",
    "Version": "v1.2.0",
    "Env":     "Production",
})
```

### Option 2: Type-Safe Function Template

```go
type ErrorData struct {
    Service string
    Code    int
    Message string
}

// 1. Register a typed function
notify.RegisterFunc("error_alert", func(data ErrorData) *message.Card {
    return message.NewCard().
        SetStatus(message.StatusDanger).
        SetTitle(fmt.Sprintf("❌ Error in %s", data.Service)).
        AddSection(
            message.NewSection().
                AddField("Code", fmt.Sprintf("%d", data.Code)).
                AddField("Detail", data.Message),
        )
})

// 2. Send fast with your struct
_ = sender.SendTemplate(ctx, "error_alert", ErrorData{
    Service: "auth-service",
    Code:    500,
    Message: "Database connection failed",
})
```

---

## ⚙️ Configuration & Switching Providers

You can change providers easily using a `Config` struct (from YAML, JSON, or environment variables):

```go
cfg := notify.Config{
    Provider:  "lark", // or "gchat", "broadcast", "fallback"
    Endpoint:  "https://open.larksuite.com/open-apis/bot/v2/hook/xxx",
    Secret:    "your-lark-secret", // optional HMAC secret for Lark
    Timeout:   "5s",
    Retries:   3,                  // auto retry 3 times on error
    RateLimit: 5.0,                // max 5 requests per second
}

sender, err := notify.NewFromConfig(cfg)
```

---

## 🛡️ Middlewares (Retry, Rate Limit, Logging)

You can add helper tools to your sender pipeline:

```go
sender := notify.New(baseSender,
    notify.WithMiddleware(
        middleware.Logging(slog.Default()),                       // logs info and errors
        middleware.RateLimit(middleware.NewRateLimiter(5.0, 10)), // 5 req/s
        middleware.Retry(3),                                      // retry 3 times
    ),
)
```

---

## 🔀 Multi-Channel (Broadcast & Fallback)

### Send to Both Google Chat and Lark (Broadcast)

```go
cfg := notify.Config{
    Provider: notify.ProviderBroadcast,
    Children: []notify.Config{
        {Provider: "gchat", Endpoint: "https://chat.googleapis.com/..."},
        {Provider: "lark", Endpoint: "https://open.larksuite.com/..."},
    },
}

sender, _ := notify.NewFromConfig(cfg)
_ = sender.Send(ctx, card.Wrap()) // sends to both platforms at the same time
```

### Fallback to Secondary Channel on Error

```go
cfg := notify.Config{
    Provider: notify.ProviderFallback,
    Children: []notify.Config{
        {Provider: "lark", Endpoint: "https://primary-lark-url"},
        {Provider: "gchat", Endpoint: "https://secondary-gchat-url"},
    },
}

sender, _ := notify.NewFromConfig(cfg)
_ = sender.Send(ctx, card.Wrap()) // tries Lark first; if it fails, sends to GChat
```

---

## 🧪 Unit Testing in Your Application

You do not need to send real webhooks in your unit tests. Use `testutil.MockSender`:

```go
func TestMyService(t *testing.T) {
    mock := testutil.NewMockSender()
    recorder := testutil.NewRecorder(mock)

    // Pass mock to your service
    myService := NewOrderService(mock)
    err := myService.CancelOrder(context.Background(), "ORD-123")

    // Check the results
    if err != nil {
        t.Fatal(err)
    }
    if mock.SentCount() != 1 {
        t.Errorf("expected 1 message, got %d", mock.SentCount())
    }
    if !recorder.HasTitle("Order Cancelled") {
        t.Errorf("title does not match")
    }
}
```

---

## 📁 Project Structure

```
go-notify/
├── notifier.go                 # Main entry, Sender interface, New()
├── options.go                  # Configuration options
├── config.go                   # Config struct (JSON / YAML / Viper)
│
├── message/                    # Message models (Card, Section, Button, Status)
├── template/                   # Template registry and render engine
├── provider/
│   ├── gchat/                  # Google Chat webhook adapter
│   └── lark/                   # Lark / Feishu webhook adapter
├── middleware/                 # Retry, Rate Limiter, and Logging
├── router/                     # Broadcast and Fallback routing
├── testutil/                   # MockSender and Recorder for unit tests
└── examples/                   # Working code examples
```

---

## 📄 License

[MIT License](LICENSE)
