package terraform

// FakeClient is a test implementation of Client that returns a static version.
type FakeClient struct{}

func NewFakeClient() *FakeClient {
	return &FakeClient{}
}

func (f *FakeClient) GetLatestVersion() (string, error) {
	return "1.14.6", nil
}

var _ Client = (*FakeClient)(nil)
