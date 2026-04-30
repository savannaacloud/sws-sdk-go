package sws

import "context"

// DatabaseService manages managed-database instances (mysql, postgresql, …).
type DatabaseService struct {
	client *Client
}

func (s *DatabaseService) ListInstances(ctx context.Context) ([]Database, error) {
	var out []Database
	if err := s.client.do(ctx, "GET", "/api/database/instances", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *DatabaseService) GetInstance(ctx context.Context, id string) (*Database, error) {
	var out Database
	if err := s.client.do(ctx, "GET", "/api/database/instances/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateInstance provisions a new managed database. AdminPassword is
// required; AdminUser defaults to "admin" server-side when blank.
func (s *DatabaseService) CreateInstance(ctx context.Context, opts *CreateDatabaseOpts) (*Database, error) {
	var out Database
	if err := s.client.do(ctx, "POST", "/api/database/instances", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *DatabaseService) DeleteInstance(ctx context.Context, id string) error {
	return s.client.do(ctx, "DELETE", "/api/database/instances/"+id, nil, nil)
}
