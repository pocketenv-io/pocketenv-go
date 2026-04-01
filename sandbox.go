package pocketenv

import (
	"context"
	"net/url"
)

type Sandbox struct {
	SandboxView
	client *Client
}

func (s *Sandbox) Refresh(ctx context.Context) error {
	updated, err := s.client.GetSandbox(ctx, s.ID)
	if err != nil {
		return err
	}
	s.SandboxView = updated.SandboxView
	return nil
}

func (s *Sandbox) Start(ctx context.Context) error {
	return s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.startSandbox", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Sandbox) Stop(ctx context.Context) error {
	return s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.stopSandbox", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Sandbox) Delete(ctx context.Context) error {
	return s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.deleteSandbox", url.Values{"id": {s.ID}}, nil, nil)
}

func (s *Sandbox) Exec(ctx context.Context, command string) (*ExecResult, error) {
	var result ExecResult
	err := s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.exec", url.Values{"id": {s.ID}}, map[string]string{"command": command}, &result)
	return &result, err
}

func (s *Sandbox) ExposePort(ctx context.Context, port int, description string) error {
	body := map[string]any{"port": port}
	if description != "" {
		body["description"] = description
	}
	return s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.exposePort", url.Values{"id": {s.ID}}, body, nil)
}

func (s *Sandbox) UnexposePort(ctx context.Context, port int) error {
	return s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.unexposePort", url.Values{"id": {s.ID}}, map[string]int{"port": port}, nil)
}

func (s *Sandbox) GetExposedPorts(ctx context.Context) ([]ExposedPort, error) {
	var result struct {
		Ports []ExposedPort `json:"ports"`
	}
	err := s.client.get(ctx, "/xrpc/io.pocketenv.sandbox.getExposedPorts", url.Values{"id": {s.ID}}, &result)
	return result.Ports, err
}

func (s *Sandbox) GetSshKeys(ctx context.Context) (*SshKeysView, error) {
	var result struct {
		SshKeys SshKeysView `json:"sshKeys"`
	}
	if err := s.client.get(ctx, "/xrpc/io.pocketenv.sandbox.getSshKeys", url.Values{"id": {s.ID}}, &result); err != nil {
		return nil, err
	}
	return &result.SshKeys, nil
}

func (s *Sandbox) PutSshKeys(ctx context.Context, publicKey, privateKey string) error {
	return s.client.post(ctx, "/xrpc/io.pocketenv.sandbox.putSshKeys", url.Values{"id": {s.ID}}, map[string]string{
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
