package run

// Run accepts a workspace, and executes the provided command in it, uploading outputs to the correct place, afterwards.
//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=run_mock.go -source=run.go -package=run
type Run interface {
	// something, something.
}
