package infratestsv1

type TestSandboxRequest struct {
	SandboxName string `json:"sandbox_name"`
}

func (req *TestSandboxRequest) Validate() error {
	return nil
}

type TestSandboxResponse struct {
	SandboxName string `json:"sandbox_name"`
}

func (req *TestSandboxResponse) Validate() error {
	return nil
}
