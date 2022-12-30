package s3

import "context"

// GetKey assumes a role and fetches the given key
func (r *repo) GetOrgsKey(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (r *repo) GetDeploymentsKey(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

func (r *repo) GetInstallationsKey(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
