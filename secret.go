package pocketenv

import (
	"context"
	"fmt"
	"net/url"
)

type SecretClient struct {
	client    *Client
	sandboxID string
}

func (sc *SecretClient) Add(ctx context.Context, name, value string) (*SecretView, error) {
	var result SecretView
	body := map[string]any{
		"secret": map[string]string{
			"sandboxId": sc.sandboxID,
			"name":      name,
			"value":     value,
		},
	}
	if err := sc.client.post(ctx, "/xrpc/io.pocketenv.secret.addSecret", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (sc *SecretClient) List(ctx context.Context, offset, limit int) ([]SecretView, int, error) {
	var result struct {
		Secrets []SecretView `json:"secrets"`
		Total   int          `json:"total"`
	}
	params := url.Values{
		"sandboxId": {sc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := sc.client.get(ctx, "/xrpc/io.pocketenv.secret.getSecrets", params, &result); err != nil {
		return nil, 0, err
	}
	return result.Secrets, result.Total, nil
}

func (sc *SecretClient) Get(ctx context.Context, id string) (*SecretView, error) {
	var result SecretView
	if err := sc.client.get(ctx, "/xrpc/io.pocketenv.secret.getSecret", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (sc *SecretClient) Update(ctx context.Context, id, name, value string) (*SecretView, error) {
	var result SecretView
	body := map[string]any{
		"id": id,
		"secret": map[string]string{
			"sandboxId": sc.sandboxID,
			"name":      name,
			"value":     value,
		},
	}
	if err := sc.client.post(ctx, "/xrpc/io.pocketenv.secret.updateSecret", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (sc *SecretClient) Delete(ctx context.Context, id string) error {
	return sc.client.post(ctx, "/xrpc/io.pocketenv.secret.deleteSecret", url.Values{"id": {id}}, nil, nil)
}
