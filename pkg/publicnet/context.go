package publicnet

import "context"

type publicOnlyKey struct{}

// WithPublicOnly marks ctx as running on the control plane with destinations
// that come from untrusted input. Shared clients that also run in the
// customer's cloud, where private destinations are legitimate, read the mark
// through PublicOnly and dial only public addresses when it is set.
func WithPublicOnly(ctx context.Context) context.Context {
	return context.WithValue(ctx, publicOnlyKey{}, true)
}

func PublicOnly(ctx context.Context) bool {
	marked, _ := ctx.Value(publicOnlyKey{}).(bool)
	return marked
}
