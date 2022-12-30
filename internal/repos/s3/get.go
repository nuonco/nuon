package s3

import "context"

// GetKey assumes a role and fetches the given key
func (r *repo) GetKey(ctx context.Context, bucket, key string, roleCfg RoleConfig) ([]byte, error) {
	return nil, nil
}
