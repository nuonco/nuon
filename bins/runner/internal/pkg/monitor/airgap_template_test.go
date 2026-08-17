package monitor

import (
	"strings"
	"testing"
)

func TestRunnerServiceAirgapAWSUsesPackagedRunner(t *testing.T) {
	if !strings.Contains(RunnerServiceAirgapAWS, "--volume /opt/nuon/runner/bin/runner:/bin/runner:ro") {
		t.Fatal("air-gap runner service must execute the packaged runner binary")
	}
}
