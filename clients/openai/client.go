// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package openai provides a client for the chat completions and models
// endpoints of the OpenAI API, as implemented by OpenAI compatible servers
// such as oMLX, llama-server, Ollama and vLLM, as well as hosted services.
// It is implemented using cloudeng.io/webapi/operations, whose options,
// eg. for rate control and logging, may be specified using
// WithOperationsOptions.
//
// The client only sends requests and returns the model's responses: it
// never executes the tool calls that a model requests, which is left to
// the caller.
//
// The API key is sent only to the configured server: redirects are not
// followed, and sending a key over plain http to a host other than
// localhost is an error unless WithAllowInsecure is specified. The key is
// never included in an error.
package openai

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"cloudeng.io/webapi/operations"
)

const (
	// DefaultMaxResponseBytes is the maximum size of a response that is
	// not streamed, and of each event in a streamed response.
	DefaultMaxResponseBytes = 32 << 20

	authorizationHeader = "Authorization"
	eventStream         = "text/event-stream"
)

type options struct {
	httpClient    *http.Client
	auth          operations.Auth
	apiKey        string
	authHeader    string
	model         string
	timeout       time.Duration
	allowInsecure bool
	maxResponse   int64
	opOpts        []operations.Option
}

// Option represents an option to New.
type Option func(*options)

// WithHTTPClient specifies the http.Client to use, http.DefaultClient's
// transport is used by default. The Client uses a copy of hc that does
// not follow redirects. hc must not have a Timeout since that would
// include the time spent reading a streamed response, use WithTimeout
// instead.
func WithHTTPClient(hc *http.Client) Option {
	return func(o *options) {
		o.httpClient = hc
	}
}

// WithAPIKey specifies the API key to send with each request, see
// WithAuthHeader and APIKey.
func WithAPIKey(key string) Option {
	return func(o *options) {
		o.apiKey = key
	}
}

// WithAuthHeader specifies the HTTP header that carries the API key
// specified by WithAPIKey, see APIKey.
func WithAuthHeader(header string) Option {
	return func(o *options) {
		o.authHeader = header
	}
}

// WithAuth specifies the operations.Auth used to authorize requests, eg.
// one that obtains a key using cloudeng.io/webapi/operations/apitokens.
// It takes precedence over WithAPIKey.
func WithAuth(auth operations.Auth) Option {
	return func(o *options) {
		o.auth = auth
	}
}

// WithModel specifies the model to use for requests that do not specify
// one.
func WithModel(model string) Option {
	return func(o *options) {
		o.model = model
	}
}

// WithTimeout specifies the time allowed for each request, including the
// reading of a streamed response. There is no timeout by default, other
// than that of the context.
func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.timeout = d
	}
}

// WithAllowInsecure allows the API key to be sent over plain http to a
// host other than localhost.
func WithAllowInsecure(allow bool) Option {
	return func(o *options) {
		o.allowInsecure = allow
	}
}

// WithMaxResponseBytes specifies the maximum size of a response that is not
// streamed, and of each event in a streamed response, see
// DefaultMaxResponseBytes. A response that is too large results in an error
// that wraps operations.ErrResponseTooLarge.
func WithMaxResponseBytes(n int64) Option {
	return func(o *options) {
		o.maxResponse = n
	}
}

// WithOperationsOptions specifies options for the operations endpoints
// used to issue requests, eg. operations.WithRateController or
// operations.WithLogger. Options that specify the http.Client,
// authorization or the encoding of requests and responses are
// overridden by the Client.
func WithOperationsOptions(opts ...operations.Option) Option {
	return func(o *options) {
		o.opOpts = append(o.opOpts, opts...)
	}
}

// APIKey implements operations.Auth for an API key. A Header of
// Authorization, or an empty Header, carries the key as 'Bearer <key>';
// any other header, eg. api-key or x-api-key, carries it as is.
type APIKey struct {
	Header string
	Key    string
}

// WithAuthorization implements operations.Auth.
func (k APIKey) WithAuthorization(_ context.Context, req *http.Request) error {
	header := cmp.Or(k.Header, authorizationHeader)
	if strings.EqualFold(header, authorizationHeader) {
		req.Header.Set(authorizationHeader, "Bearer "+k.Key)
		return nil
	}
	req.Header.Set(header, k.Key)
	return nil
}

// Client is a client for an OpenAI compatible API.
type Client struct {
	opts    options
	baseURL *url.URL
	models  *operations.Endpoint[modelList]
	chat    *operations.PutEndpoint[ChatRequest, ChatResponse]
	stream  *operations.Endpoint[any]
}

type modelList struct {
	Data []Model `json:"data"`
}

// New returns a Client for the API at baseURL, which includes the API's
// version prefix, eg. https://api.openai.com/v1 or
// http://127.0.0.1:8000/v1.
func New(baseURL string, opts ...Option) (*Client, error) {
	c := &Client{}
	c.opts.maxResponse = DefaultMaxResponseBytes
	for _, fn := range opts {
		fn(&c.opts)
	}
	u, err := parseURL(baseURL)
	if err != nil {
		return nil, err
	}
	if c.opts.auth == nil && c.opts.apiKey != "" {
		c.opts.auth = APIKey{Header: c.opts.authHeader, Key: c.opts.apiKey}
	}
	if c.opts.auth != nil && u.Scheme == "http" && !isLoopback(u.Hostname()) && !c.opts.allowInsecure {
		return nil, fmt.Errorf("openai: %v: an API key would be sent over plain http to a host other than localhost, use https or allow insecure connections", u)
	}
	c.baseURL = u
	epOpts := slices.Concat(c.opts.opOpts, []operations.Option{
		operations.WithHTTPClient(c.httpClient()),
		operations.WithAuth(c.opts.auth),
		operations.WithUnmarshal(json.Unmarshal, operations.JSONEncoding),
		operations.WithMarshal(json.Marshal, operations.JSONEncoding),
		operations.WithSuccessCodes(http.StatusOK),
		operations.WithMaxResponseBytes(c.opts.maxResponse),
	})
	c.models = operations.NewEndpoint[modelList](epOpts...)
	c.chat = operations.NewPutEndpoint[ChatRequest, ChatResponse](epOpts...)
	c.stream = operations.NewStream(epOpts...)
	return c, nil
}

func parseURL(baseURL string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("openai: %v: the scheme must be http or https", u.Redacted())
	}
	if u.Host == "" {
		return nil, fmt.Errorf("openai: %v: no host", u.Redacted())
	}
	if u.User != nil {
		return nil, fmt.Errorf("openai: %v: credentials must not be included in the url", u.Redacted())
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u, nil
}

// httpClient returns a copy of the configured http.Client that does not
// follow redirects.
func (c *Client) httpClient() *http.Client {
	hc := http.Client{}
	if c.opts.httpClient != nil {
		hc = *c.opts.httpClient
	}
	hc.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &hc
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// BaseURL returns the base URL of the API.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}

// Model returns the default model, see WithModel.
func (c *Client) Model() string {
	return c.opts.model
}

func (c *Client) url(path string) string {
	u := *c.baseURL
	u.Path += path
	return u.String()
}

// Models returns the models listed by the server.
func (c *Client) Models(ctx context.Context) ([]Model, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	models, body, _, err := c.models.Get(ctx, c.url("/models"))
	if err != nil {
		return nil, c.error(err, body, http.MethodGet, "/models")
	}
	return models.Data, nil
}

// Chat returns the completion for req.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ctx, cancel := c.withTimeout(ctx)
	defer cancel()
	req.Stream, req.StreamOptions = false, nil
	if err := c.setModel(&req); err != nil {
		return nil, err
	}
	resp, body, _, err := c.chat.Post(ctx, c.url(chatPath), req)
	if err != nil {
		return nil, c.error(err, body, http.MethodPost, chatPath)
	}
	return &resp, nil
}

const chatPath = "/chat/completions"

func (c *Client) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.opts.timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.opts.timeout)
}

func (c *Client) setModel(req *ChatRequest) error {
	req.Model = cmp.Or(req.Model, c.opts.model)
	if req.Model == "" {
		return errors.New("openai: no model specified")
	}
	return nil
}

// postStream issues a streamed chat request and returns the response,
// whose body is unread, if its status is 200.
func (c *Client) postStream(ctx context.Context, req ChatRequest) (*http.Response, error) {
	if err := c.setModel(&req); err != nil {
		return nil, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(chatPath), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Accept", eventStream)
	resp, err := c.stream.IssueStream(ctx, hreq)
	if err != nil {
		var errBody []byte
		if resp != nil && resp.Body != nil {
			errBody, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
		return nil, c.error(err, errBody, http.MethodPost, chatPath)
	}
	return resp, nil
}

// error returns an *APIError for an unsuccessful response, and err,
// annotated with the request, otherwise.
func (c *Client) error(err error, body []byte, method, path string) error {
	var oe *operations.Error
	if errors.As(err, &oe) && oe.StatusCode != 0 && oe.StatusCode != http.StatusOK {
		ae := newAPIError(body)
		ae.StatusCode, ae.Method, ae.Path, ae.Err = oe.StatusCode, method, path, oe
		return ae
	}
	return fmt.Errorf("openai: %v %v: %w", method, path, err)
}
