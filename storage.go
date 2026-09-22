package sws

import "context"

// StorageService manages block-storage volumes.
type StorageService struct {
	client *Client
}

func (s *StorageService) ListVolumes(ctx context.Context) ([]Volume, error) {
	var out []Volume
	if err := s.client.do(ctx, "GET", "/api/v1/block-storage/volumes", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *StorageService) GetVolume(ctx context.Context, id string) (*Volume, error) {
	var out Volume
	if err := s.client.do(ctx, "GET", "/api/v1/block-storage/volumes/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *StorageService) CreateVolume(ctx context.Context, opts *CreateVolumeOpts) (*Volume, error) {
	var out Volume
	if err := s.client.do(ctx, "POST", "/api/v1/block-storage/volumes", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *StorageService) DeleteVolume(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/api/v1/block-storage/volumes/"+id, nil, nil)
}

func (s *StorageService) AttachVolume(ctx context.Context, volumeID, instanceID string) error {
	return s.client.do(ctx, "POST", "/api/v1/block-storage/volumes/"+volumeID+"/attach",
		map[string]string{"instance_id": instanceID}, nil)
}

func (s *StorageService) DetachVolume(ctx context.Context, volumeID string) error {
	return s.client.do(ctx, "POST", "/api/v1/block-storage/volumes/"+volumeID+"/detach", nil, nil)
}
