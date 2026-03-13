package monitor

import (
	"testing"
)

func TestParseCPUSampleLine(t *testing.T) {
	t.Parallel()

	sample, err := parseCPUSampleLine("cpu  4705 0 2288 104589 99 0 248 0 0 0")
	if err != nil {
		t.Fatalf("parseCPUSampleLine returned error: %v", err)
	}

	if sample.idle != 104688 {
		t.Fatalf("unexpected idle value: got %d, want %d", sample.idle, 104688)
	}

	if sample.total != 111929 {
		t.Fatalf("unexpected total value: got %d, want %d", sample.total, 111929)
	}
}

func TestParseMemInfo(t *testing.T) {
	t.Parallel()

	content := "MemTotal:       16348624 kB\nMemFree:         789912 kB\nMemAvailable:    9215332 kB\n"

	total, available, err := parseMemInfo(content)
	if err != nil {
		t.Fatalf("parseMemInfo returned error: %v", err)
	}

	if total != 16348624 {
		t.Fatalf("unexpected total memory: got %d, want %d", total, 16348624)
	}

	if available != 9215332 {
		t.Fatalf("unexpected available memory: got %d, want %d", available, 9215332)
	}
}

func TestParseDockerStatsOutput(t *testing.T) {
	t.Parallel()

	out := "abc123\trunner-a\t17.50%\t3.50%\t512.0MiB / 8.00GiB\nxyz987\trunner-b\t4.00%\t1.25%\t256MiB / 8.00GiB\n"

	stats, err := parseDockerStatsOutput(out)
	if err != nil {
		t.Fatalf("parseDockerStatsOutput returned error: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("unexpected stat count: got %d, want %d", len(stats), 2)
	}

	if stats[0].ContainerID != "abc123" {
		t.Fatalf("unexpected first container id: got %q", stats[0].ContainerID)
	}

	if stats[0].ContainerName != "runner-a" {
		t.Fatalf("unexpected first container name: got %q", stats[0].ContainerName)
	}

	if stats[0].CPUUtilizationPct != 17.5 {
		t.Fatalf("unexpected first container cpu pct: got %f, want %f", stats[0].CPUUtilizationPct, 17.5)
	}

	if stats[0].MemoryUtilization != 3.5 {
		t.Fatalf("unexpected first container memory pct: got %f, want %f", stats[0].MemoryUtilization, 3.5)
	}

	if stats[0].MemoryUsedBytes <= 0 {
		t.Fatalf("expected used bytes to be > 0, got %f", stats[0].MemoryUsedBytes)
	}

	if stats[0].MemoryLimitBytes <= stats[0].MemoryUsedBytes {
		t.Fatalf("expected memory limit to be > used bytes, got used=%f limit=%f", stats[0].MemoryUsedBytes, stats[0].MemoryLimitBytes)
	}
}

func TestParseByteSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		expect  float64
		wantErr bool
	}{
		{name: "bytes", input: "10B", expect: 10},
		{name: "decimal mb", input: "1.5MB", expect: 1.5e6},
		{name: "binary mib", input: "2MiB", expect: 2 * 1024 * 1024},
		{name: "empty", input: "", expect: 0},
		{name: "unknown unit", input: "4XB", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseByteSize(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseByteSize returned error for %q: %v", tc.input, err)
			}

			if actual != tc.expect {
				t.Fatalf("unexpected value for %q: got %f, want %f", tc.input, actual, tc.expect)
			}
		})
	}
}
