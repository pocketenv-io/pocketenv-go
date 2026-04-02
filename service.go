package pocketenv

import (
	"net/url"
)

type ServiceClient struct {
	client    *Client
	sandboxID string
}

func (sc *ServiceClient) Add(input AddServiceInput) error {
	body := map[string]any{
		"service": input,
	}
	return sc.client.post("/xrpc/io.pocketenv.service.addService", url.Values{"sandboxId": {sc.sandboxID}}, body, nil)
}

func (sc *ServiceClient) List() ([]ServiceView, error) {
	var result struct {
		Services []ServiceView `json:"services"`
	}
	if err := sc.client.get("/xrpc/io.pocketenv.service.getServices", url.Values{"sandboxId": {sc.sandboxID}}, &result); err != nil {
		return nil, err
	}
	return result.Services, nil
}

func (sc *ServiceClient) Update(serviceID string, input UpdateServiceInput) error {
	body := map[string]any{
		"service": input,
	}
	return sc.client.post("/xrpc/io.pocketenv.service.updateService", url.Values{"serviceId": {serviceID}}, body, nil)
}

func (sc *ServiceClient) Delete(serviceID string) error {
	return sc.client.post("/xrpc/io.pocketenv.service.deleteService", url.Values{"serviceId": {serviceID}}, nil, nil)
}

func (sc *ServiceClient) Start(serviceID string) error {
	return sc.client.post("/xrpc/io.pocketenv.service.startService", url.Values{"serviceId": {serviceID}}, nil, nil)
}

func (sc *ServiceClient) Stop(serviceID string) error {
	return sc.client.post("/xrpc/io.pocketenv.service.stopService", url.Values{"serviceId": {serviceID}}, nil, nil)
}

func (sc *ServiceClient) Restart(serviceID string) error {
	return sc.client.post("/xrpc/io.pocketenv.service.restartService", url.Values{"serviceId": {serviceID}}, nil, nil)
}
