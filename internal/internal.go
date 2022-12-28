package internal

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_internal_test.go -source=internal.go -package=internal
type TestInterface interface {
	GetService() error
}
