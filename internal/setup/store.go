package setup

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"sg-supervisor/internal/config"
)

const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusSkipped   = "skipped"

	FieldLicense     = "license"
	FieldTelegramBot = "telegram-bot"
)

type Field struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Required  bool   `json:"required"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type State struct {
	Version       int     `json:"version"`
	Fields        []Field `json:"fields"`
	LastUpdatedAt string  `json:"lastUpdatedAt,omitempty"`
}

type Summary struct {
	Complete       bool
	BlockingFields []string
	Required       []Field
	Optional       []Field
}

type Store struct {
	path string
}

func NewStore(layout config.Layout) *Store {
	return &Store{path: filepath.Join(layout.ConfigDir, "setup-state.json")}
}

func (s *Store) Ensure(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return s.save(defaultState())
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Store) Load(ctx context.Context) (State, error) {
	if err := s.Ensure(ctx); err != nil {
		return State{}, err
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, err
	}
	if len(state.Fields) == 0 {
		state = defaultState()
	}
	return state, nil
}

func (s *Store) UpdateField(ctx context.Context, key, status string) (State, error) {
	state, err := s.Load(ctx)
	if err != nil {
		return State{}, err
	}
	if status != StatusPending && status != StatusCompleted && status != StatusSkipped {
		return State{}, errors.New("setup field status must be pending, completed, or skipped")
	}

	updated := false
	for index := range state.Fields {
		field := &state.Fields[index]
		if field.Key != key {
			continue
		}
		if field.Required && status == StatusSkipped {
			return State{}, errors.New("required setup fields cannot be skipped")
		}
		field.Status = status
		field.UpdatedAt = now()
		updated = true
	}
	if !updated {
		return State{}, errors.New("unknown setup field")
	}

	state.LastUpdatedAt = now()
	return state, s.save(state)
}

func Summarize(state State, licenseValid bool) Summary {
	required := make([]Field, 0, len(state.Fields))
	optional := make([]Field, 0, len(state.Fields))
	blocking := make([]string, 0, len(state.Fields))

	for _, field := range state.Fields {
		current := field
		if current.Key == FieldLicense {
			if licenseValid {
				current.Status = StatusCompleted
			} else {
				current.Status = StatusPending
			}
		}
		if current.Required {
			required = append(required, current)
			if current.Status != StatusCompleted {
				blocking = append(blocking, current.Key)
			}
			continue
		}
		optional = append(optional, current)
	}

	return Summary{
		Complete:       len(blocking) == 0,
		BlockingFields: blocking,
		Required:       required,
		Optional:       optional,
	}
}

func (s *Store) save(state State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o644)
}

func defaultState() State {
	return State{
		Version: 1,
		Fields: []Field{
			{Key: FieldLicense, Label: "License activation", Required: true, Status: StatusPending},
			{Key: FieldTelegramBot, Label: "Telegram bot configuration", Required: false, Status: StatusPending},
		},
	}
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
