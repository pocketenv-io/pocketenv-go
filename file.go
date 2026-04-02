package pocketenv

import (
	"fmt"
	"net/url"
)

type FileClient struct {
	client    *Client
	sandboxID string
}

func (fc *FileClient) Add(path, content string) error {
	encrypted, err := fc.client.encrypt(content)
	if err != nil {
		return err
	}
	body := map[string]any{
		"file": map[string]string{
			"sandboxId": fc.sandboxID,
			"path":      path,
			"content":   encrypted,
		},
	}
	return fc.client.post("/xrpc/io.pocketenv.file.addFile", nil, body, nil)
}

func (fc *FileClient) List(offset, limit int) ([]FileView, int, error) {
	var result struct {
		Files []FileView `json:"files"`
		Total int        `json:"total"`
	}
	params := url.Values{
		"sandboxId": {fc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := fc.client.get("/xrpc/io.pocketenv.file.getFiles", params, &result); err != nil {
		return nil, 0, err
	}
	return result.Files, result.Total, nil
}

func (fc *FileClient) Get(id string) (*FileView, error) {
	var result struct {
		File FileView `json:"file"`
	}
	if err := fc.client.get("/xrpc/io.pocketenv.file.getFile", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &result.File, nil
}

func (fc *FileClient) Update(id, path, content string) error {
	encrypted, err := fc.client.encrypt(content)
	if err != nil {
		return err
	}
	body := map[string]any{
		"id": id,
		"file": map[string]string{
			"path":    path,
			"content": encrypted,
		},
	}
	return fc.client.post("/xrpc/io.pocketenv.file.updateFile", nil, body, nil)
}

func (fc *FileClient) Delete(id string) error {
	return fc.client.post("/xrpc/io.pocketenv.file.deleteFile", url.Values{"id": {id}}, nil, nil)
}
