package monitor

import (
	"strings"
	"testing"
)

func TestRunnerServiceCustomerManagedAWSUsesPackagedRunner(t *testing.T) {
	if !strings.Contains(RunnerServiceCustomerManagedAWS, "--volume /opt/nuon/runner/bin/runner:/bin/runner:ro") {
		t.Fatal("offline runner service must execute the packaged runner binary")
	}
}
