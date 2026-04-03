package pocketenv

import (
	"fmt"
	"net/url"
)

type Secret struct {
	secretView
	sandboxID string
	client    *Client
}

func (s *Secret) Update(name, value string) (*Secret, error) {
	encrypted, err := s.client.encrypt(value)
	if err != nil {
		return nil, err
	}
	var result secretView
	body := map[string]any{
		"id": s.ID,
		"secret": map[string]string{
			"sandboxId": s.sandboxID,
			"name":      name,
			"value":     encrypted,
		},
	}
	if err := s.client.post("/xrpc/io.pocketenv.secret.updateSecret", nil, body, &result); err != nil {
		return nil, err
	}
	s.secretView = result
	return s, nil
}

func (s *Secret) Delete() error {
	return s.client.post("/xrpc/io.pocketenv.secret.deleteSecret", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Secret) Refresh() error {
	var result secretView
	if err := s.client.get("/xrpc/io.pocketenv.secret.getSecret", url.Values{"id": {s.ID}}, &result); err != nil {
		return err
	}
	s.secretView = result
	return nil
}

type SecretClient struct {
	client    *Client
	sandboxID string
}

func (sc *SecretClient) Create(name, value string) (*Secret, error) {
	encrypted, err := sc.client.encrypt(value)
	if err != nil {
		return nil, err
	}
	var result secretView
	body := map[string]any{
		"secret": map[string]string{
			"sandboxId": sc.sandboxID,
			"name":      name,
			"value":     encrypted,
		},
	}
	if err := sc.client.post("/xrpc/io.pocketenv.secret.addSecret", nil, body, &result); err != nil {
		return nil, err
	}
	return &Secret{secretView: result, sandboxID: sc.sandboxID, client: sc.client}, nil
}

func (sc *SecretClient) List(offset, limit int) (Page[*Secret], error) {
	var raw struct {
		Secrets []secretView `json:"secrets"`
		Total   int          `json:"total"`
	}
	params := url.Values{
		"sandboxId": {sc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := sc.client.get("/xrpc/io.pocketenv.secret.getSecrets", params, &raw); err != nil {
		return Page[*Secret]{}, err
	}
	items := make([]*Secret, len(raw.Secrets))
	for i := range raw.Secrets {
		items[i] = &Secret{secretView: raw.Secrets[i], sandboxID: sc.sandboxID, client: sc.client}
	}
	return Page[*Secret]{Items: items, Total: raw.Total}, nil
}

func (sc *SecretClient) Get(id string) (*Secret, error) {
	var result secretView
	if err := sc.client.get("/xrpc/io.pocketenv.secret.getSecret", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &Secret{secretView: result, sandboxID: sc.sandboxID, client: sc.client}, nil
}
