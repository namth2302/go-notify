package template_test

import (
	"sync"
	"testing"

	"github.com/namth2302/go-notify/message"
	"github.com/namth2302/go-notify/template"
)

type DeployAlertData struct {
	Service string
	Version string
	Env     string
}

func TestDeclarativeTemplateRender(t *testing.T) {
	fieldTpl := template.NewFieldTemplate("Arch", "arm64").SetShort(false)

	tpl := template.NewCardTemplate().
		SetStatus(message.StatusSuccess).
		SetTitle("🚀 Deploy: {{ .Service }}").
		SetSubtitle("Env: {{ .Env }}").
		AddSection(
			template.NewSectionTemplate().
				SetHeader("Info").
				SetText("Deployed by CI/CD").
				AddField("Version", "{{ .Version }}").
				AddFields(fieldTpl).
				SetImage("https://example.com/{{ .Service }}.png", "Logo").
				AddButton(template.NewButtonTemplate("Details", "https://ci.internal/{{ .Service }}")),
		).
		AddAction(
			template.NewButtonTemplate("View Prod", "https://prod.internal").AsPrimary(),
			template.NewButtonTemplate("Rollback", "https://rollback.internal").AsDanger(),
		)

	data := DeployAlertData{
		Service: "payment-api",
		Version: "v1.2.3",
		Env:     "production",
	}

	msg, err := tpl.Render(data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	cardMsg, ok := msg.(*message.CardMessage)
	if !ok {
		t.Fatalf("expected *message.CardMessage, got %T", msg)
	}

	card := cardMsg.Card
	if card.Title != "🚀 Deploy: payment-api" {
		t.Errorf("unexpected title: %q", card.Title)
	}
	if card.Subtitle != "Env: production" {
		t.Errorf("unexpected subtitle: %q", card.Subtitle)
	}
	if len(card.Sections) != 1 || card.Sections[0].Fields[0].Value != "v1.2.3" {
		t.Errorf("unexpected section field: %+v", card.Sections)
	}
	if card.Sections[0].Image.URL != "https://example.com/payment-api.png" {
		t.Errorf("unexpected image url: %q", card.Sections[0].Image.URL)
	}
	if card.Actions[0].URL != "https://prod.internal" {
		t.Errorf("unexpected action url: %q", card.Actions[0].URL)
	}
	if card.Actions[1].Type != message.ButtonTypeDanger {
		t.Errorf("expected button type danger, got %v", card.Actions[1].Type)
	}
}

func TestFunctionalTemplateRender(t *testing.T) {
	type ErrorAlert struct {
		ErrCode int
		Message string
	}

	fnTpl := template.FromCardFunc(func(data ErrorAlert) *message.Card {
		status := message.StatusWarning
		if data.ErrCode >= 500 {
			status = message.StatusDanger
		}
		return message.NewCard().
			SetStatus(status).
			SetTitle("Error Alert").
			AddSection(message.NewSection().AddField("Error", data.Message))
	})

	msg, err := fnTpl.Render(ErrorAlert{ErrCode: 502, Message: "Bad Gateway"})
	if err != nil {
		t.Fatalf("functional render failed: %v", err)
	}
	card := msg.(*message.CardMessage).Card
	if card.Status != message.StatusDanger {
		t.Errorf("expected StatusDanger, got %v", card.Status)
	}

	// Test wrong type
	_, err = fnTpl.Render("invalid type")
	if err == nil {
		t.Errorf("expected type mismatch error, got nil")
	}

	// Test nil function
	nilTpl := template.NewFuncTemplate(nil)
	_, err = nilTpl.Render(nil)
	if err == nil {
		t.Errorf("expected error for nil func")
	}

	// Test func returning nil card
	nilCardTpl := template.FromCardFunc(func(s string) *message.Card {
		return nil
	})
	_, err = nilCardTpl.Render("test")
	if err == nil {
		t.Errorf("expected error for nil card return")
	}
}

func TestRegistryConcurrency(t *testing.T) {
	reg := template.NewRegistry()

	// Register template
	err := reg.Register("simple", template.NewCardTemplate().SetTitle("Hello {{ .Name }}"))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg, err := reg.Render("simple", map[string]string{"Name": "Tester"})
			if err != nil {
				t.Errorf("concurrent render failed: %v", err)
				return
			}
			card := msg.(*message.CardMessage).Card
			if card.Title != "Hello Tester" {
				t.Errorf("unexpected title: %q", card.Title)
			}
		}(i)
	}
	wg.Wait()

	// Test not found
	_, err = reg.Render("non_existent", nil)
	if err == nil {
		t.Errorf("expected not found error")
	}

	// Test invalid registration
	if err := reg.Register("", nil); err == nil {
		t.Errorf("expected empty name error")
	}
	if err := reg.Register("foo", nil); err == nil {
		t.Errorf("expected nil template error")
	}
}

func TestGlobalRegistry(t *testing.T) {
	err := template.Register("global_decl", template.NewCardTemplate().SetTitle("Title {{ .A }}"))
	if err != nil {
		t.Fatalf("global register failed: %v", err)
	}
	if _, ok := template.Get("global_decl"); !ok {
		t.Errorf("expected global template to be found")
	}

	err = template.RegisterFunc("global_test", func(data string) *message.Card {
		return message.NewCard().SetTitle("Hello " + data)
	})
	if err != nil {
		t.Fatalf("global register failed: %v", err)
	}

	msg, err := template.Render("global_test", "World")
	if err != nil {
		t.Fatalf("global render failed: %v", err)
	}
	card := msg.(*message.CardMessage).Card
	if card.Title != "Hello World" {
		t.Errorf("unexpected title: %q", card.Title)
	}

	if template.DefaultRegistry() == nil {
		t.Errorf("expected default registry")
	}
}
