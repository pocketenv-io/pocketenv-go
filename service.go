package pocketenv

import (
	"net/url"
)

type Service struct {
	serviceView
	sandboxID string
	client    *Client
}

func (s *Service) Update(input UpdateServiceInput) error {
	return s.client.post("/xrpc/io.pocketenv.service.updateService", url.Values{"serviceId": {s.ID}}, map[string]any{"service": input}, nil)
}

func (s *Service) Delete() error {
	return s.client.post("/xrpc/io.pocketenv.service.deleteService", url.Values{"serviceId": {s.ID}}, nil, nil)
}

func (s *Service) Start() error {
	return s.client.post("/xrpc/io.pocketenv.service.startService", url.Values{"serviceId": {s.ID}}, nil, nil)
}

func (s *Service) Stop() error {
	return s.client.post("/xrpc/io.pocketenv.service.stopService", url.Values{"serviceId": {s.ID}}, nil, nil)
}

func (s *Service) Restart() error {
	return s.client.post("/xrpc/io.pocketenv.service.restartService", url.Values{"serviceId": {s.ID}}, nil, nil)
}

type ServiceClient struct {
	client    *Client
	sandboxID string
}

func (sc *ServiceClient) Create(input AddServiceInput) error {
	return sc.client.post("/xrpc/io.pocketenv.service.addService", url.Values{"sandboxId": {sc.sandboxID}}, map[string]any{"service": input}, nil)
}

func (sc *ServiceClient) List() ([]*Service, error) {
	var raw struct {
		Services []serviceView `json:"services"`
	}
	if err := sc.client.get("/xrpc/io.pocketenv.service.getServices", url.Values{"sandboxId": {sc.sandboxID}}, &raw); err != nil {
		return nil, err
	}
	services := make([]*Service, len(raw.Services))
	for i := range raw.Services {
		services[i] = &Service{serviceView: raw.Services[i], sandboxID: sc.sandboxID, client: sc.client}
	}
	return services, nil
}
