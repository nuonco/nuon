package installdelegationdns

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=client_mock.go -source=client.go -package=installdelegationdns

//go:generate -command temporal-gen go run github.com/powertoolsdev/mono/pkg/gen/temporal-gen
//go:generate temporal-gen
