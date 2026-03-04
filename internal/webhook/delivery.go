package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Service struct {
	store  store.Store
	client *http.Client
}

func New(st store.Store) *Service {
	return &Service{
		store:  st,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

func (s *Service) Deliver(ctx context.Context, namespaceID, eventType string, payload any) error {
	hooks, err := s.store.ListWebhooks(ctx, namespaceID)
	if err != nil {
		return err
	}
	ev := Event{Type: eventType, Timestamp: time.Now().UTC(), Payload: payload}
	body, _ := json.Marshal(ev)
	for _, hook := range hooks {
		if !hook.Enabled || !containsEvent(hook.Events, eventType) {
			continue
		}
		_ = s.sendWithRetry(ctx, hook, body)
	}
	return nil
}

func (s *Service) sendWithRetry(ctx context.Context, hook domain.Webhook, body []byte) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}
		if err := s.send(ctx, hook, body); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (s *Service) send(ctx context.Context, hook domain.Webhook, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TaskForge-Signature", sign(body, hook.Secret))
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return domainErr("webhook delivery failed")
	}
	return nil
}

func sign(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func containsEvent(events []string, target string) bool {
	for _, e := range events {
		if e == target || e == "*" {
			return true
		}
	}
	return false
}

func domainErr(msg string) error {
	return &whError{msg: msg}
}

type whError struct{ msg string }

func (e *whError) Error() string { return e.msg }
