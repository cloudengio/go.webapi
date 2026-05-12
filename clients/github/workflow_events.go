// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Repository represents the repository information included in webhook payloads.
type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Private  bool   `json:"private"`
	Owner    Actor  `json:"owner"`
}

// WorkflowJobEvent is the payload delivered to a webhook for workflow_job events.
// Action is one of "queued", "in_progress", or "completed".
type WorkflowJobEvent struct {
	Action      string     `json:"action"`
	WorkflowJob Job        `json:"workflow_job"`
	Repository  Repository `json:"repository"`
	Sender      Actor      `json:"sender"`
}

// WorkflowRunEvent is the payload delivered to a webhook for workflow_run events.
// Action is one of "requested", "in_progress", or "completed".
type WorkflowRunEvent struct {
	Action      string      `json:"action"`
	WorkflowRun WorkflowRun `json:"workflow_run"`
	Repository  Repository  `json:"repository"`
	Sender      Actor       `json:"sender"`
}

// MockWebhook creates signed HTTP POST requests that mimic GitHub webhook
// deliveries. It is intended for testing webhook relays and handlers.
type MockWebhook struct {
	secret string
	owner  string
	repo   string
}

// NewMockWebhook returns a MockWebhook for the given owner/repo. secret is the
// webhook secret used to produce X-Hub-Signature-256 headers; pass an empty
// string to skip signing.
func NewMockWebhook(owner, repo, secret string) *MockWebhook {
	return &MockWebhook{secret: secret, owner: owner, repo: repo}
}

// JobRequest returns a signed HTTP POST request to targetURL for a workflow_job
// event. action must be one of "queued", "in_progress", or "completed".
func (m *MockWebhook) JobRequest(ctx context.Context, targetURL, action string, job Job) (*http.Request, error) {
	event := WorkflowJobEvent{
		Action:      action,
		WorkflowJob: job,
		Repository:  m.repository(),
		Sender:      Actor{Login: m.owner},
	}
	return m.newRequest(ctx, targetURL, "workflow_job", event)
}

// RunRequest returns a signed HTTP POST request to targetURL for a workflow_run
// event. action must be one of "requested", "in_progress", or "completed".
func (m *MockWebhook) RunRequest(ctx context.Context, targetURL, action string, run WorkflowRun) (*http.Request, error) {
	event := WorkflowRunEvent{
		Action:      action,
		WorkflowRun: run,
		Repository:  m.repository(),
		Sender:      Actor{Login: m.owner},
	}
	return m.newRequest(ctx, targetURL, "workflow_run", event)
}

func (m *MockWebhook) repository() Repository {
	return Repository{
		Name:     m.repo,
		FullName: m.owner + "/" + m.repo,
		HTMLURL:  "https://github.com/" + url.PathEscape(m.owner) + "/" + url.PathEscape(m.repo),
		Owner:    Actor{Login: m.owner},
	}
}

func (m *MockWebhook) newRequest(ctx context.Context, targetURL, eventType string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", eventType)
	req.Header.Set("X-GitHub-Delivery", webhookDeliveryID())
	if m.secret != "" {
		mac := hmac.New(sha256.New, []byte(m.secret))
		mac.Write(body)
		req.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	return req, nil
}

// VerifyWebhookSignature reports whether the X-Hub-Signature-256 header value
// matches the HMAC-SHA256 of body computed with secret. This is the check a
// relay or handler performs on receipt.
func VerifyWebhookSignature(secret string, body []byte, signature string) bool {
	sig, ok := strings.CutPrefix(signature, "sha256=")
	if !ok {
		return false
	}
	want, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), want)
}

// MockJob returns a Job populated with typical values for use in tests.
// Callers may overwrite any field before passing it to MockWebhook.JobRequest.
func MockJob(owner, repo string) Job {
	now := time.Now().UTC()
	return Job{
		ID:           1,
		RunID:        1,
		Name:         "test-job",
		Status:       "queued",
		HeadBranch:   "main",
		HeadSHA:      strings.Repeat("a", 40),
		WorkflowName: "CI",
		HTMLURL:      fmt.Sprintf("https://github.com/%s/%s/actions/runs/1/jobs/1", url.PathEscape(owner), url.PathEscape(repo)),
		StartedAt:    &now,
	}
}

// MockRun returns a WorkflowRun populated with typical values for use in tests.
// Callers may overwrite any field before passing it to MockWebhook.RunRequest.
func MockRun(owner, repo string) WorkflowRun {
	now := time.Now().UTC()
	return WorkflowRun{
		ID:           1,
		Name:         "CI",
		HeadBranch:   "main",
		HeadSHA:      strings.Repeat("a", 40),
		RunNumber:    1,
		RunAttempt:   1,
		Status:       "requested",
		Event:        "push",
		WorkflowID:   1,
		WorkflowName: "CI",
		URL:          fmt.Sprintf("%s/repos/%s/%s/actions/runs/1", APIHost, url.PathEscape(owner), url.PathEscape(repo)),
		HTMLURL:      fmt.Sprintf("https://github.com/%s/%s/actions/runs/1", url.PathEscape(owner), url.PathEscape(repo)),
		CreatedAt:    &now,
		Actor:        Actor{Login: owner},
	}
}

// webhookDeliveryID returns a random UUID v4-shaped string matching the format
// GitHub uses for X-GitHub-Delivery headers.
func webhookDeliveryID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
