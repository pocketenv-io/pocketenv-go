package pocketenv

import (
	"fmt"
	"net/url"
)

type SecretClient struct {
	client    *Client
	sandboxID string
}

func (sc *SecretClient) Add(name, value string) (*SecretView, error) {
	encrypted, err := sc.client.encrypt(value)
	if err != nil {
		return nil, err
	}
	var result SecretView
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
	return &result, nil
}

func (sc *SecretClient) List(offset, limit int) ([]SecretView, int, error) {
	var result struct {
		Secrets []SecretView `json:"secrets"`
		Total   int          `json:"total"`
	}
	params := url.Values{
		"sandboxId": {sc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := sc.client.get("/xrpc/io.pocketenv.secret.getSecrets", params, &result); err != nil {
		return nil, 0, err
	}
	return result.Secrets, result.Total, nil
}

func (sc *SecretClient) Get(id string) (*SecretView, error) {
	var result SecretView
	if err := sc.client.get("/xrpc/io.pocketenv.secret.getSecret", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (sc *SecretClient) Update(id, name, value string) (*SecretView, error) {
	encrypted, err := sc.client.encrypt(value)
	if err != nil {
		return nil, err
	}
	var result SecretView
	body := map[string]any{
		"id": id,
		"secret": map[string]string{
			"sandboxId": sc.sandboxID,
			"name":      name,
			"value":     encrypted,
		},
	}
	if err := sc.client.post("/xrpc/io.pocketenv.secret.updateSecret", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (sc *SecretClient) Delete(id string) error {
	return sc.client.post("/xrpc/io.pocketenv.secret.deleteSecret", url.Values{"id": {id}}, nil, nil)
}
