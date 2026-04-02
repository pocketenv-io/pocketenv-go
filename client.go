package pocketenv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const defaultBaseURL = "https://api.pocketenv.io"

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Sandbox(id string) *Sandbox {
	return &Sandbox{SandboxView: SandboxView{ID: id}, client: c}
}

// CreateSandbox creates a new sandbox and returns it as a ready-to-use handle.
func (c *Client) CreateSandbox(input CreateSandboxInput) (*Sandbox, error) {
	var view SandboxView
	if err := c.post("/xrpc/io.pocketenv.sandbox.createSandbox", nil, input, &view); err != nil {
		return nil, err
	}
	return &Sandbox{SandboxView: view, client: c}, nil
}

func (c *Client) GetSandbox(id string) (*Sandbox, error) {
	var result struct {
		Sandbox SandboxView `json:"sandbox"`
	}
	if err := c.get("/xrpc/io.pocketenv.sandbox.getSandbox", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &Sandbox{SandboxView: result.Sandbox, client: c}, nil
}

func (c *Client) ListSandboxes(offset, limit int) ([]*Sandbox, int, error) {
	var result struct {
		Sandboxes []SandboxView `json:"sandboxes"`
		Total     int           `json:"total"`
	}
	params := url.Values{
		"offset": {fmt.Sprintf("%d", offset)},
		"limit":  {fmt.Sprintf("%d", limit)},
	}
	if err := c.get("/xrpc/io.pocketenv.sandbox.getSandboxes", params, &result); err != nil {
		return nil, 0, err
	}
	sandboxes := make([]*Sandbox, len(result.Sandboxes))
	for i, v := range result.Sandboxes {
		sandboxes[i] = &Sandbox{SandboxView: v, client: c}
	}
	return sandboxes, result.Total, nil
}

func (c *Client) Secrets(sandboxID string) *SecretClient {
	return &SecretClient{client: c, sandboxID: sandboxID}
}

func (c *Client) Variables(sandboxID string) *VariableClient {
	return &VariableClient{client: c, sandboxID: sandboxID}
}

func (c *Client) Files(sandboxID string) *FileClient {
	return &FileClient{client: c, sandboxID: sandboxID}
}

func (c *Client) Volumes(sandboxID string) *VolumeClient {
	return &VolumeClient{client: c, sandboxID: sandboxID}
}

func (c *Client) Services(sandboxID string) *ServiceClient {
	return &ServiceClient{client: c, sandboxID: sandboxID}
}

// HTTP helpers

func (c *Client) get(path string, params url.Values, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *Client) post(path string, params url.Values, body any, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pocketenv: HTTP %d: %s", resp.StatusCode, string(b))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
