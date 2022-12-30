package waypoint

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=waypoint_mocks.go -source=waypoint.go -package=waypoint
type Repo interface{}

type repo struct{}

var _ Repo = (*repo)(nil)
