package pocketenv

import (
	"fmt"
	"net/url"
)

type Volume struct {
	volumeView
	sandboxID string
	client    *Client
}

func (v *Volume) Update(name, path string) error {
	return v.client.post("/xrpc/io.pocketenv.volume.updateVolume", nil, map[string]any{
		"id": v.ID,
		"volume": map[string]string{
			"sandboxId": v.sandboxID,
			"name":      name,
			"path":      path,
		},
	}, nil)
}

func (v *Volume) Delete() error {
	return v.client.post("/xrpc/io.pocketenv.volume.deleteVolume", url.Values{"id": {v.ID}}, nil, nil)
}

func (v *Volume) Refresh() error {
	var raw struct {
		Volume volumeView `json:"volume"`
	}
	if err := v.client.get("/xrpc/io.pocketenv.volume.getVolume", url.Values{"id": {v.ID}}, &raw); err != nil {
		return err
	}
	v.volumeView = raw.Volume
	return nil
}

type VolumeClient struct {
	client    *Client
	sandboxID string
}

func (vc *VolumeClient) Create(name, path string) error {
	return vc.client.post("/xrpc/io.pocketenv.volume.addVolume", nil, map[string]any{
		"volume": map[string]string{
			"sandboxId": vc.sandboxID,
			"name":      name,
			"path":      path,
		},
	}, nil)
}

func (vc *VolumeClient) List(offset, limit int) (Page[*Volume], error) {
	var raw struct {
		Volumes []volumeView `json:"volumes"`
		Total   int          `json:"total"`
	}
	params := url.Values{
		"sandboxId": {vc.sandboxID},
		"offset":    {fmt.Sprintf("%d", offset)},
		"limit":     {fmt.Sprintf("%d", limit)},
	}
	if err := vc.client.get("/xrpc/io.pocketenv.volume.getVolumes", params, &raw); err != nil {
		return Page[*Volume]{}, err
	}
	items := make([]*Volume, len(raw.Volumes))
	for i := range raw.Volumes {
		items[i] = &Volume{volumeView: raw.Volumes[i], sandboxID: vc.sandboxID, client: vc.client}
	}
	return Page[*Volume]{Items: items, Total: raw.Total}, nil
}

func (vc *VolumeClient) Get(id string) (*Volume, error) {
	var raw struct {
		Volume volumeView `json:"volume"`
	}
	if err := vc.client.get("/xrpc/io.pocketenv.volume.getVolume", url.Values{"id": {id}}, &raw); err != nil {
		return nil, err
	}
	return &Volume{volumeView: raw.Volume, sandboxID: vc.sandboxID, client: vc.client}, nil
}
