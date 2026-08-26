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

	gogithub "github.com/google/go-github/v89/github"
)

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
func (m *MockWebhook) JobRequest(ctx context.Context, targetURL, action string, job *gogithub.WorkflowJob) (*http.Request, error) {
	event := gogithub.WorkflowJobEvent{
		Action:      new(action),
		WorkflowJob: job,
		Repo:        m.repository(),
		Sender:      &gogithub.User{Login: new(m.owner)},
	}
	return m.newRequest(ctx, targetURL, "workflow_job", event)
}

// RunRequest returns a signed HTTP POST request to targetURL for a workflow_run
// event. action must be one of "requested", "in_progress", or "completed".
func (m *MockWebhook) RunRequest(ctx context.Context, targetURL, action string, run *gogithub.WorkflowRun) (*http.Request, error) {
	event := gogithub.WorkflowRunEvent{
		Action:      new(action),
		WorkflowRun: run,
		Repo:        m.repository(),
		Sender:      &gogithub.User{Login: new(m.owner)},
	}
	return m.newRequest(ctx, targetURL, "workflow_run", event)
}

func (m *MockWebhook) repository() *gogithub.Repository {
	return &gogithub.Repository{
		Name:     new(m.repo),
		FullName: new(m.owner + "/" + m.repo),
		HTMLURL:  new("https://github.com/" + url.PathEscape(m.owner) + "/" + url.PathEscape(m.repo)),
		Owner:    &gogithub.User{Login: new(m.owner)},
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

// MockJob returns a WorkflowJob populated with typical values for use in tests.
// Callers may overwrite any field before passing it to MockWebhook.JobRequest.
func MockJob(owner, repo string) *gogithub.WorkflowJob {
	now := gogithub.Timestamp{Time: time.Now().UTC()}
	return &gogithub.WorkflowJob{
		ID:           new(int64(1)),
		RunID:        new(int64(1)),
		Name:         new("test-job"),
		Status:       new("queued"),
		HeadBranch:   new("main"),
		HeadSHA:      new(strings.Repeat("a", 40)),
		WorkflowName: new("CI"),
		HTMLURL:      new(fmt.Sprintf("https://github.com/%s/%s/actions/runs/1/jobs/1", url.PathEscape(owner), url.PathEscape(repo))),
		StartedAt:    &now,
	}
}

// MockRun returns a WorkflowRun populated with typical values for use in tests.
// Callers may overwrite any field before passing it to MockWebhook.RunRequest.
func MockRun(owner, repo string) *gogithub.WorkflowRun {
	now := gogithub.Timestamp{Time: time.Now().UTC()}
	return &gogithub.WorkflowRun{
		ID:         new(int64(1)),
		Name:       new("CI"),
		HeadBranch: new("main"),
		HeadSHA:    new(strings.Repeat("a", 40)),
		RunNumber:  new(1),
		RunAttempt: new(1),
		Status:     new("requested"),
		Event:      new("push"),
		WorkflowID: new(int64(1)),
		URL:        new(fmt.Sprintf("%s/repos/%s/%s/actions/runs/1", APIHost, url.PathEscape(owner), url.PathEscape(repo))),
		HTMLURL:    new(fmt.Sprintf("https://github.com/%s/%s/actions/runs/1", url.PathEscape(owner), url.PathEscape(repo))),
		CreatedAt:  &now,
		Actor:      &gogithub.User{Login: new(owner)},
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
