package pocketenv

import (
	"fmt"
	"net/url"
)

type VolumeClient struct {
	client    *Client
	sandboxID string
}

func (vc *VolumeClient) Add(name, path string) error {
	body := map[string]any{
		"volume": map[string]string{
			"sandboxId": vc.sandboxID,
			"name":      name,
			"path":      path,
		},
	}
	return vc.client.post("/xrpc/io.pocketenv.volume.addVolume", nil, body, nil)
}

func (vc *VolumeClient) List(offset, limit int) ([]VolumeView, int, error) {
	var result struct {
		Volumes []VolumeView `json:"volumes"`
		Total   int          `json:"total"`
	}
	params := url.Values{
		"sandboxId": {vc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := vc.client.get("/xrpc/io.pocketenv.volume.getVolumes", params, &result); err != nil {
		return nil, 0, err
	}
	return result.Volumes, result.Total, nil
}

func (vc *VolumeClient) Get(id string) (*VolumeView, error) {
	var result struct {
		Volume VolumeView `json:"volume"`
	}
	if err := vc.client.get("/xrpc/io.pocketenv.volume.getVolume", url.Values{"id": {id}}, &result); err != nil {
		return nil, err
	}
	return &result.Volume, nil
}

func (vc *VolumeClient) Update(id, name, path string) error {
	body := map[string]any{
		"id": id,
		"volume": map[string]string{
			"sandboxId": vc.sandboxID,
			"name":      name,
			"path":      path,
		},
	}
	return vc.client.post("/xrpc/io.pocketenv.volume.updateVolume", nil, body, nil)
}

func (vc *VolumeClient) Delete(id string) error {
	return vc.client.post("/xrpc/io.pocketenv.volume.deleteVolume", url.Values{"id": {id}}, nil, nil)
}
