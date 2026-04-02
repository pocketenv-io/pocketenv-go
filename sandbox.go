package pocketenv

import (
	"net/url"
	"strings"
)

type Sandbox struct {
	SandboxView
	client *Client
}

func (s *Sandbox) Refresh() error {
	updated, err := s.client.GetSandbox(s.ID)
	if err != nil {
		return err
	}
	s.SandboxView = updated.SandboxView
	return nil
}

func (s *Sandbox) Start() error {
	return s.client.post("/xrpc/io.pocketenv.sandbox.startSandbox", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Sandbox) Stop() error {
	return s.client.post("/xrpc/io.pocketenv.sandbox.stopSandbox", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Sandbox) Delete() error {
	return s.client.post("/xrpc/io.pocketenv.sandbox.deleteSandbox", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Sandbox) Exec(command ...string) (*ExecResult, error) {
	var result ExecResult
	err := s.client.post("/xrpc/io.pocketenv.sandbox.exec", url.Values{"id": {s.ID}}, map[string]string{"command": strings.Join(command, " ")}, &result)
	return &result, err
}

func (s *Sandbox) ExposePort(port int, description string) error {
	body := map[string]any{"port": port}
	if description != "" {
		body["description"] = description
	}
	return s.client.post("/xrpc/io.pocketenv.sandbox.exposePort", url.Values{"id": {s.ID}}, body, nil)
}

func (s *Sandbox) UnexposePort(port int) error {
	return s.client.post("/xrpc/io.pocketenv.sandbox.unexposePort", url.Values{"id": {s.ID}}, map[string]int{"port": port}, nil)
}

func (s *Sandbox) GetExposedPorts() ([]ExposedPort, error) {
	var result struct {
		Ports []ExposedPort `json:"ports"`
	}
	err := s.client.get("/xrpc/io.pocketenv.sandbox.getExposedPorts", url.Values{"id": {s.ID}}, &result)
	return result.Ports, err
}

func (s *Sandbox) GetSshKeys() (*SshKeysView, error) {
	var result struct {
		SshKeys SshKeysView `json:"sshKeys"`
	}
	if err := s.client.get("/xrpc/io.pocketenv.sandbox.getSshKeys", url.Values{"id": {s.ID}}, &result); err != nil {
		return nil, err
	}
	return &result.SshKeys, nil
}

func (s *Sandbox) PutSshKeys(publicKey, privateKey string) error {
	return s.client.post("/xrpc/io.pocketenv.sandbox.putSshKeys", url.Values{"id": {s.ID}}, map[string]string{
		"publicKey":  publicKey,
		"privateKey": privateKey,
	}, nil)
}

func (s *Sandbox) Secrets() *SecretClient {
	return s.client.Secrets(s.ID)
}

func (s *Sandbox) Variables() *VariableClient {
	return s.client.Variables(s.ID)
}

func (s *Sandbox) Files() *FileClient {
	return s.client.Files(s.ID)
}

func (s *Sandbox) Volumes() *VolumeClient {
	return s.client.Volumes(s.ID)
}

func (s *Sandbox) Services() *ServiceClient {
	return s.client.Services(s.ID)
}
