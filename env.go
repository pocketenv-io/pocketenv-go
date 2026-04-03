package pocketenv

import (
	"fmt"
	"net/url"
)

type Variable struct {
	variableView
	sandboxID string
	client    *Client
}

func (v *Variable) Update(name, value string) (*Variable, error) {
	var result variableView
	body := map[string]any{
		"id": v.ID,
		"variable": map[string]string{
			"sandboxId": v.sandboxID,
			"name":      name,
			"value":     value,
		},
	}
	if err := v.client.post("/xrpc/io.pocketenv.variable.updateVariable", nil, body, &result); err != nil {
		return nil, err
	}
	v.variableView = result
	return v, nil
}

func (v *Variable) Delete() error {
	return v.client.post("/xrpc/io.pocketenv.variable.deleteVariable", url.Values{"id": {v.ID}}, nil, nil)
}

func (v *Variable) Refresh() error {
	var result variableView
	if err := v.client.get("/xrpc/io.pocketenv.variable.getVariable", url.Values{"id": {v.ID}}, &result); err != nil {
		return err
	}
	v.variableView = result
	return nil
}

type VariableClient struct {
	client    *Client
	sandboxID string
}

func (vc *VariableClient) Create(name, value string) (*Variable, error) {
	var result variableView
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
	return &Variable{variableView: result, sandboxID: vc.sandboxID, client: vc.client}, nil
}

func (vc *VariableClient) List(offset, limit int) (Page[*Variable], error) {
	var raw struct {
		Variables []variableView `json:"variables"`
		Total     int            `json:"total"`
	}
	params := url.Values{
		"sandboxId": {vc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := vc.client.get("/xrpc/io.pocketenv.variable.getVariables", params, &raw); err != nil {
		return Page[*Variable]{}, err
	}
	items := make([]*Variable, len(raw.Variables))
	for i := range raw.Variables {
		items[i] = &Variable{variableView: raw.Variables[i], sandboxID: vc.sandboxID, client: vc.client}
	}
	return Page[*Variable]{Items: items, Total: raw.Total}, nil
}

func (vc *VariableClient) Get(id string) (*Variable, error) {
	var result variableView
	if err := vc.client.get("/xrpc/io.pocketenv.variable.getVariable", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &Variable{variableView: result, sandboxID: vc.sandboxID, client: vc.client}, nil
}
