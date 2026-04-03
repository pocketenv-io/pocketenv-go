package pocketenv

import (
	"fmt"
	"net/url"
)

type File struct {
	fileView
	sandboxID string
	client    *Client
}

func (f *File) Update(path, content string) error {
	encrypted, err := f.client.encrypt(content)
	if err != nil {
		return err
	}
	return f.client.post("/xrpc/io.pocketenv.file.updateFile", nil, map[string]any{
		"id": f.ID,
		"file": map[string]string{
			"path":    path,
			"content": encrypted,
		},
	}, nil)
}

func (f *File) Delete() error {
	return f.client.post("/xrpc/io.pocketenv.file.deleteFile", url.Values{"id": {f.ID}}, nil, nil)
}

func (f *File) Refresh() error {
	var raw struct {
		File fileView `json:"file"`
	}
	if err := f.client.get("/xrpc/io.pocketenv.file.getFile", url.Values{"id": {f.ID}}, &raw); err != nil {
		return err
	}
	f.fileView = raw.File
	return nil
}

type FileClient struct {
	client    *Client
	sandboxID string
}

func (fc *FileClient) Create(path, content string) error {
	encrypted, err := fc.client.encrypt(content)
	if err != nil {
		return err
	}
	return fc.client.post("/xrpc/io.pocketenv.file.addFile", nil, map[string]any{
		"file": map[string]string{
			"sandboxId": fc.sandboxID,
			"path":      path,
			"content":   encrypted,
		},
	}, nil)
}

func (fc *FileClient) List(offset, limit int) (Page[*File], error) {
	var raw struct {
		Files []fileView `json:"files"`
		Total int        `json:"total"`
	}
	params := url.Values{
		"sandboxId": {fc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := fc.client.get("/xrpc/io.pocketenv.file.getFiles", params, &raw); err != nil {
		return Page[*File]{}, err
	}
	items := make([]*File, len(raw.Files))
	for i := range raw.Files {
		items[i] = &File{fileView: raw.Files[i], sandboxID: fc.sandboxID, client: fc.client}
	}
	return Page[*File]{Items: items, Total: raw.Total}, nil
}

func (fc *FileClient) Get(id string) (*File, error) {
	var raw struct {
		File fileView `json:"file"`
	}
	if err := fc.client.get("/xrpc/io.pocketenv.file.getFile", url.Values{"id": {id}}, &raw); err != nil {
		return nil, err
	}
	return &File{fileView: raw.File, sandboxID: fc.sandboxID, client: fc.client}, nil
}
