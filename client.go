package pocketenv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const defaultBaseURL = "https://api.pocketenv.io"

type Client struct {
	baseURL    string
	token      string
	publicKey  string
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

func WithPublicKey(publicKey string) Option {
	return func(c *Client) {
		c.publicKey = publicKey
	}
}

func New(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.publicKey == "" {
		if pk := os.Getenv("POCKETENV_PUBLIC_KEY"); pk != "" {
			c.publicKey = pk
		} else {
			c.publicKey = defaultPublicKey
		}
	}
	if c.token == "" {
		if token := os.Getenv("POCKETENV_TOKEN"); token != "" {
			c.token = token
		} else {
			token, err := readTokenFile()
			if err != nil {
				return nil, fmt.Errorf("pocketenv: no token provided, POCKETENV_TOKEN not set, and ~/.pocketenv/token.json is unavailable: %w", err)
			}
			if token == "" {
				return nil, fmt.Errorf("pocketenv: token is empty in ~/.pocketenv/token.json")
			}
			c.token = token
		}
	}
	return c, nil
}

func readTokenFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(home, ".pocketenv", "token.json"))
	if err != nil {
		return "", err
	}
	var v struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	return v.Token, nil
}

// SandboxRef returns a lightweight handle for a known sandbox ID without
// making an API call. Use GetSandbox to fetch its current data from the API.
func (c *Client) SandboxRef(id string) *Sandbox {
	return &Sandbox{sandboxView: sandboxView{ID: id}, client: c}
}

func (c *Client) CreateSandbox(input CreateSandboxInput) (*Sandbox, error) {
	var view sandboxView
	if err := c.post("/xrpc/io.pocketenv.sandbox.createSandbox", nil, input, &view); err != nil {
		return nil, err
	}
	return &Sandbox{sandboxView: view, client: c}, nil
}

func (c *Client) GetSandbox(id string) (*Sandbox, error) {
	var result struct {
		Sandbox sandboxView `json:"sandbox"`
	}
	if err := c.get("/xrpc/io.pocketenv.sandbox.getSandbox", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &Sandbox{sandboxView: result.Sandbox, client: c}, nil
}

func (c *Client) ListSandboxes(offset, limit int) (Page[*Sandbox], error) {
	var result struct {
		Sandboxes []sandboxView `json:"sandboxes"`
		Total     int           `json:"total"`
	}
	params := url.Values{
		"offset": {fmt.Sprintf("%d", offset)},
		"limit":  {fmt.Sprintf("%d", limit)},
	}
	if err := c.get("/xrpc/io.pocketenv.sandbox.getSandboxes", params, &result); err != nil {
		return Page[*Sandbox]{}, err
	}
	items := make([]*Sandbox, len(result.Sandboxes))
	for i, v := range result.Sandboxes {
		items[i] = &Sandbox{sandboxView: v, client: c}
	}
	return Page[*Sandbox]{Items: items, Total: result.Total}, nil
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

func (c *Client) encrypt(value string) (string, error) {
	return sealedBoxEncrypt(c.publicKey, value)
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
		return &Error{StatusCode: resp.StatusCode, Message: string(b)}
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
