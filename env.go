package pocketenv

import (
	"fmt"
	"net/url"
)

type VariableClient struct {
	client    *Client
	sandboxID string
}

func (vc *VariableClient) Add(name, value string) (*VariableView, error) {
	var result VariableView
	body := map[string]any{
		"variable": map[string]string{
			"sandboxId": vc.sandboxID,
			"name":      name,
			"value":     value,
		},
	}
	if err := vc.client.post("/xrpc/io.pocketenv.variable.addVariable", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (vc *VariableClient) List(offset, limit int) ([]VariableView, int, error) {
	var result struct {
		Variables []VariableView `json:"variables"`
		Total     int            `json:"total"`
	}
	params := url.Values{
		"sandboxId": {vc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := vc.client.get("/xrpc/io.pocketenv.variable.getVariables", params, &result); err != nil {
		return nil, 0, err
	}
	return result.Variables, result.Total, nil
}

func (vc *VariableClient) Get(id string) (*VariableView, error) {
	var result VariableView
	if err := vc.client.get("/xrpc/io.pocketenv.variable.getVariable", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (vc *VariableClient) Update(id, name, value string) (*VariableView, error) {
	var result VariableView
	body := map[string]any{
		"id": id,
		"variable": map[string]string{
			"sandboxId": vc.sandboxID,
			"name":      name,
			"value":     value,
		},
	}
	if err := vc.client.post("/xrpc/io.pocketenv.variable.updateVariable", nil, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (vc *VariableClient) Delete(id string) error {
	return vc.client.post("/xrpc/io.pocketenv.variable.deleteVariable", url.Values{"id": {id}}, nil, nil)
}
