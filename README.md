# Go Notify (Google Chat & Lark)

[![Go Reference](https://pkg.go.dev/badge/github.com/namth2302/go-notify.svg)](https://pkg.go.dev/github.com/namth2302/go-notify)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Zero Dependencies](https://img.shields.io/badge/Dependencies-Zero-brightgreen.svg)](go.mod)

Thư viện Go chuyên nghiệp dùng chung (Reusable Library) để gửi thông báo / cảnh báo đến **Google Chat** và **Lark (Feishu)** qua Webhook.

---

## ✨ Điểm Nổi Bật (Key Features)

- **Zero External Dependencies**: 100% Go Standard Library (`net/http`, `crypto/hmac`, `log/slog`, `text/template`). Hoàn toàn không làm phình `go.mod` của các dự án import.
- **Clean Architecture & Outside-In Design**: Phân tách rõ ràng giữa Domain Models, Adapters, Pipeline Interceptors và Router.
- **Universal Card DSL**: Viết cấu trúc thông báo 1 lần (`message.NewCard()`), tự động render sang chuẩn **Google Chat CardV2** và **Lark Interactive Card 2.0**.
- **Template Registry (Gửi Nhanh)**: Hỗ trợ đăng ký Template trước (cả Declarative `{{ .Var }}` và Type-Safe Generics) rồi gửi nhanh chỉ bằng 1 dòng code.
- **Fast Provider Switch**: Chuyển đổi giữa Google Chat, Lark, Broadcast hoặc Fallback chỉ bằng thay đổi biến môi trường / Config struct.
- **Resilience & Middleware**: Tích hợp sẵn Retry (Exponential Backoff + Jitter), Token-Bucket Rate Limiter, `log/slog` logging.
- **Testability First**: Cung cấp sẵn package `testutil` (`MockSender`, `Recorder`) để các service viết Unit Test mà không cần gọi API thật.

---

## 🚀 Cài Đặt (Installation)

```bash
go get github.com/namth2302/go-notify
```

---

## 📦 Hướng Dẫn Sử Dụng Nhanh (Quickstart)

### 1. Gửi Card Universal (Tự động thích ứng GChat & Lark)

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

    // Khởi tạo từ biến môi trường (NOTIFY_PROVIDER=gchat hoặc NOTIFY_PROVIDER=lark)
    sender, err := notify.NewFromEnv()
    if err != nil {
        panic(err)
    }

    // Tạo Universal Card
    card := message.NewCard().
        SetStatus(message.StatusDanger). // Tự động map màu đỏ (Lark/GChat)
        SetTitle("🚨 [Alert] High Memory Usage").
        SetSubtitle("Service: payment-api | Env: Production").
        AddSection(
            message.NewSection().
                AddField("Memory", "94.2%").
                AddField("Threshold", "85%").
                AddField("Host", "k8s-node-01"),
        ).
        AddAction(
            message.NewButton("Xem Grafana", "https://grafana.internal/d/123"),
            message.NewButton("Xác nhận", "https://ops.internal/ack").AsPrimary(),
        )

    // Gửi đi
    _, _ = sender.Send(ctx, card.Wrap())
}
```

---

## 📑 Đăng Ký Template & Gửi Nhanh (Template Registry)

### Cách 1: Declarative Template (Dùng placeholder `{{ .Var }}`)

```go
// 1. Đăng ký template lúc khởi động App
notify.RegisterTemplate("deploy_success", template.NewCardTemplate().
    SetStatus(message.StatusSuccess).
    SetTitle("🚀 Deploy Thành Công: {{ .ServiceName }}").
    AddSection(
        template.NewSectionTemplate().
            AddField("Version", "{{ .Version }}").
            AddField("Environment", "{{ .Env }}"),
    ),
)

// 2. Gửi nhanh bất cứ lúc nào:
_ = sender.SendTemplate(ctx, "deploy_success", map[string]string{
    "ServiceName": "order-service",
    "Version":     "v1.4.2",
    "Env":         "Production",
})
```

### Cách 2: Type-Safe Functional Template (Generics)

```go
type AlertEvent struct {
    Service string
    ErrCode int
    Message string
}

// 1. Đăng ký Type-Safe template
notify.RegisterFunc("service_error", func(evt AlertEvent) *message.Card {
    return message.NewCard().
        SetStatus(message.StatusDanger).
        SetTitle(fmt.Sprintf("❌ Lỗi %d tại %s", evt.ErrCode, evt.Service)).
        AddSection(message.NewSection().AddField("Chi tiết", evt.Message))
})

// 2. Gửi nhanh với Struct:
_ = sender.SendTemplate(ctx, "service_error", AlertEvent{
    Service: "auth-api",
    ErrCode: 502,
    Message: "Bad Gateway to Redis",
})
```

---

## 🛡️ Middleware & Resilience (Khả năng chịu lỗi)

```go
sender, err := notify.NewFromConfig(cfg,
    notify.WithMiddleware(
        middleware.Logging(slog.Default()), // Log bằng slog chuẩn Go
        middleware.RateLimit(middleware.NewRateLimiter(5.0, 10)), // 5 req/s, burst 10
        middleware.Retry(3), // Tự động retry tối đa 3 lần khi gặp 429 hoặc 5xx
    ),
)
```

---

## 🔀 Multi-Channel Routing (Broadcast & Fallback)

### Broadcast (Gửi song song nhiều nền tảng)
```go
cfg := notify.Config{
    Provider: notify.ProviderBroadcast,
    Children: []notify.Config{
        {Provider: "gchat", Endpoint: "https://chat.googleapis.com/..."},
        {Provider: "lark", Endpoint: "https://open.larksuite.com/...", Secret: "xxx"},
    },
}
sender, _ := notify.NewFromConfig(cfg)
_ = sender.Send(ctx, card.Wrap()) // Gửi đồng thời tới cả GChat và Lark
```

### Fallback (Tự động chuyển kênh phụ khi kênh chính lỗi)
```go
cfg := notify.Config{
    Provider: notify.ProviderFallback,
    Children: []notify.Config{
        {Provider: "lark", Endpoint: "https://primary-lark-url"},
        {Provider: "gchat", Endpoint: "https://secondary-gchat-url"},
    },
}
sender, _ := notify.NewFromConfig(cfg)
```

---

## 🧪 Viết Unit Test Trong Dự Án Của Bạn (Testability)

Thư viện cung cấp sẵn `testutil.MockSender` và `testutil.Recorder` để bạn test nghiệp vụ mà không gửi HTTP thật:

```go
func TestOrderService(t *testing.T) {
    mock := testutil.NewMockSender()
    recorder := testutil.NewRecorder(mock)

    service := NewOrderService(mock) // Inject mock vào service
    err := service.CancelOrder(context.Background(), "ORD-123", "Khách hủy")

    assert.NoError(t, err)
    assert.Equal(t, 1, mock.SentCount())
    assert.True(t, recorder.HasTitle("❌ Đơn hàng đã hủy: ORD-123"))
    assert.True(t, recorder.HasStatus(message.StatusDanger))
}
```

---

## 📂 Cấu Trúc Dự Án (Package Layout)

```
go-notify/
├── notifier.go                 # Facade chính, Sender interface, Factory New()
├── options.go                  # Functional Options cấu hình chung
├── config.go                   # Config struct (hỗ trợ JSON/YAML/Viper)
│
├── message/                    # DOMAIN LAYER (Zero-dependency)
│   ├── message.go              # Message interface & Types (Card, Text, Raw)
│   ├── card.go                 # Card, Section, Field, Image struct & builder
│   ├── action.go               # Button, Action definitions
│   ├── status.go               # Status enum & icon helpers
│   └── result.go               # Send Result metadata
│
├── template/                   # TEMPLATE REGISTRY
│   ├── template.go             # Template interface
│   ├── declarative.go          # Declarative Card template (Go text/template)
│   ├── functional.go           # Generic type-safe functional template
│   └── registry.go             # Thread-safe template storage (sync.RWMutex)
│
├── provider/                   # ADAPTER LAYER
│   ├── gchat/                  # Google Chat Webhook (CardV2 serializer)
│   └── lark/                   # Lark / Feishu Webhook (Card 2.0 + HMAC Signer)
│
├── middleware/                 # INTERCEPTORS LAYER
│   ├── retry.go                # Exponential backoff retry with jitter
│   ├── ratelimit.go            # Token bucket rate limiter
│   └── logging.go              # slog structured logging
│
├── router/                     # MULTI-PROVIDER ROUTING
│   ├── broadcast.go            # Concurrent multi-send
│   └── fallback.go             # Sequential fallback
│
├── testutil/                   # TESTING SUITE CHO DỰ ÁN IMPORT
│   ├── mock_sender.go          # Thread-safe in-memory MockSender
│   └── recorder.go             # Helper assertions (HasTitle, HasStatus, ...)
│
└── examples/                   # SOURCE CODE MẪU
    ├── 01_simple_webhook/
    ├── 02_template_registry/
    ├── 03_switch_provider/
    └── 04_unit_test_mocking/
```

---

## 📄 Giấy Phép (License)

MIT License.
