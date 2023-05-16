package builds

import (
	"testing"
)

func TestNew(t *testing.T) {
	// TODO
	// The previous tests were only testing that the server accepted a WithService(mockSvc).
	// Since this server doesn't use a service, and we weren't testing anything else anyway,
	// I deleted them, and I'll circle back in a separate PR to write new tests.
	//
	// We'll want to start adding tests to confirm that the server, or, rather, it's handlers,
	// handle API requests correctly.
}
