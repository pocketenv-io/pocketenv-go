package pocketenv

import (
	"context"
	"fmt"
	"net/url"
)

type FileClient struct {
	client    *Client
	sandboxID string
}

func (fc *FileClient) Add(ctx context.Context, path, content string) error {
	body := map[string]any{
		"file": map[string]string{
			"sandboxId": fc.sandboxID,
			"path":      path,
			"content":   content,
		},
	}
	return fc.client.post(ctx, "/xrpc/io.pocketenv.file.addFile", nil, body, nil)
}

func (fc *FileClient) List(ctx context.Context, offset, limit int) ([]FileView, int, error) {
	var result struct {
		Files []FileView `json:"files"`
		Total int        `json:"total"`
	}
	params := url.Values{
		"sandboxId": {fc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := fc.client.get(ctx, "/xrpc/io.pocketenv.file.getFiles", params, &result); err != nil {
		return nil, 0, err
	}
	return result.Files, result.Total, nil
}

func (fc *FileClient) Get(ctx context.Context, id string) (*FileView, error) {
	var result struct {
		File FileView `json:"file"`
	}
	if err := fc.client.get(ctx, "/xrpc/io.pocketenv.file.getFile", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &result.File, nil
}

func (fc *FileClient) Update(ctx context.Context, id, path, content string) error {
	body := map[string]any{
		"id": id,
		"file": map[string]string{
			"path":    path,
			"content": content,
		},
	}
	return fc.client.post(ctx, "/xrpc/io.pocketenv.file.updateFile", nil, body, nil)
}

func (fc *FileClient) Delete(ctx context.Context, id string) error {
	return fc.client.post(ctx, "/xrpc/io.pocketenv.file.deleteFile", url.Values{"id": {id}}, nil, nil)
}
