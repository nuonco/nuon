package syncer

// Option configures optional syncer behaviour.
type Option func(*syncer)

// WithComponentBuildDispatch makes the syncer schedule component builds itself,
// the way the per-type Create*ComponentConfig handlers do: unchanged components
// reuse their existing config connection, changed components get config-created
// signals enqueued, and the scheduled set is reported through
// GetComponentsScheduled.
//
// Leave this off for callers that schedule builds in a later step of their own —
// the app branch run does this in its builds step — otherwise components build
// twice.
func WithComponentBuildDispatch() Option {
	return func(s *syncer) {
		s.dispatchBuilds = true
	}
}
