package setup

import (
	"context"
	"testing"

	"sg-supervisor/internal/config"
)

func TestLoadCreatesDefaultState(t *testing.T) {
	layout := config.NewLayout(t.TempDir())
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	store := NewStore(layout)
	state, err := store.Load(context.Background())
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(state.Fields) != 2 {
		t.Fatalf("expected 2 setup fields, got %d", len(state.Fields))
	}
}

func TestSummarizeUsesLiveLicenseValidity(t *testing.T) {
	state := defaultState()

	summary := Summarize(state, false)
	if summary.Complete {
		t.Fatalf("expected incomplete setup without license")
	}
	if len(summary.BlockingFields) != 1 || summary.BlockingFields[0] != FieldLicense {
		t.Fatalf("unexpected blocking fields: %+v", summary.BlockingFields)
	}

	summary = Summarize(state, true)
	if !summary.Complete {
		t.Fatalf("expected complete setup with valid license")
	}
	if summary.Required[0].Status != StatusCompleted {
		t.Fatalf("expected license field to be completed, got %s", summary.Required[0].Status)
	}
}

func TestUpdateFieldRejectsInvalidTransitions(t *testing.T) {
	layout := config.NewLayout(t.TempDir())
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	store := NewStore(layout)
	if _, err := store.UpdateField(context.Background(), FieldLicense, StatusSkipped); err == nil {
		t.Fatalf("expected required field skip to fail")
	}
	if _, err := store.UpdateField(context.Background(), FieldTelegramBot, StatusCompleted); err != nil {
		t.Fatalf("expected optional field update to succeed: %v", err)
	}
}
