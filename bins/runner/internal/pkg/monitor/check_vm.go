package monitor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

const (
	procStatPath = "/proc/stat"
	procMemPath  = "/proc/meminfo"

	cpuSampleInterval = 500 * time.Millisecond

	dockerStatsCmdTimeout = 10 * time.Second
	dockerStatsFormat     = "{{.Container}}\t{{.Name}}\t{{.CPUPerc}}\t{{.MemPerc}}\t{{.MemUsage}}"
)

type vmResourceStats struct {
	CPUUtilizationPct float64
	MemoryUtilization float64
	MemoryUsedBytes   float64
	MemoryTotalBytes  float64
}

type dockerResourceStat struct {
	ContainerID       string
	ContainerName     string
	CPUUtilizationPct float64
	MemoryUtilization float64
	MemoryUsedBytes   float64
	MemoryLimitBytes  float64
}

type cpuSample struct {
	total uint64
	idle  uint64
}

func (h *Monitor) checkVMResources(ctx context.Context) error {
	h.l.Debug("checking vm resources")

	vStats, err := collectVMResourceStats(ctx)
	if err != nil {
		h.l.Warn("unable to collect vm resource stats", zap.Error(err))
	} else {
		h.mw.Gauge("monitor.vm.cpu.utilization_pct", vStats.CPUUtilizationPct, nil)
		h.mw.Gauge("monitor.vm.memory.utilization_pct", vStats.MemoryUtilization, nil)
		h.mw.Gauge("monitor.vm.memory.used_bytes", vStats.MemoryUsedBytes, nil)
		h.mw.Gauge("monitor.vm.memory.total_bytes", vStats.MemoryTotalBytes, nil)

		h.l.Debug(
			"vm resource stats",
			zap.Float64("cpu_utilization_pct", vStats.CPUUtilizationPct),
			zap.Float64("memory_utilization_pct", vStats.MemoryUtilization),
			zap.Float64("memory_used_bytes", vStats.MemoryUsedBytes),
			zap.Float64("memory_total_bytes", vStats.MemoryTotalBytes),
		)
	}

	dStats, dErr := collectDockerResourceStats(ctx)
	if dErr != nil {
		h.l.Warn("unable to collect docker resource stats", zap.Error(dErr))
		return errors.Join(err, dErr)
	}

	totalCPU := 0.0
	totalMemUsed := 0.0
	totalMemLimit := 0.0

	h.mw.Gauge("monitor.docker.containers.running", float64(len(dStats)), nil)

	for _, stat := range dStats {
		tags := metrics.ToTags(map[string]string{
			"container_id":   stat.ContainerID,
			"container_name": stat.ContainerName,
		})

		h.mw.Gauge("monitor.docker.container.cpu.utilization_pct", stat.CPUUtilizationPct, tags)
		h.mw.Gauge("monitor.docker.container.memory.utilization_pct", stat.MemoryUtilization, tags)
		h.mw.Gauge("monitor.docker.container.memory.used_bytes", stat.MemoryUsedBytes, tags)
		h.mw.Gauge("monitor.docker.container.memory.limit_bytes", stat.MemoryLimitBytes, tags)

		totalCPU += stat.CPUUtilizationPct
		totalMemUsed += stat.MemoryUsedBytes
		totalMemLimit += stat.MemoryLimitBytes
	}

	h.mw.Gauge("monitor.docker.cpu.utilization_pct_total", totalCPU, nil)
	h.mw.Gauge("monitor.docker.memory.used_bytes_total", totalMemUsed, nil)
	h.mw.Gauge("monitor.docker.memory.limit_bytes_total", totalMemLimit, nil)

	h.l.Debug(
		"docker resource stats",
		zap.Int("running_containers", len(dStats)),
		zap.Float64("cpu_utilization_pct_total", totalCPU),
		zap.Float64("memory_used_bytes_total", totalMemUsed),
		zap.Float64("memory_limit_bytes_total", totalMemLimit),
	)

	if err != nil {
		return err
	}

	return nil
}

func collectVMResourceStats(ctx context.Context) (vmResourceStats, error) {
	cpuPct, err := vmCPUUtilization(ctx, cpuSampleInterval)
	if err != nil {
		return vmResourceStats{}, err
	}

	memTotalKB, memAvailableKB, err := vmMemoryInfoKB()
	if err != nil {
		return vmResourceStats{}, err
	}

	if memTotalKB == 0 {
		return vmResourceStats{}, fmt.Errorf("invalid meminfo: MemTotal is 0")
	}

	usedKB := memTotalKB - memAvailableKB
	memPct := (float64(usedKB) / float64(memTotalKB)) * 100

	return vmResourceStats{
		CPUUtilizationPct: cpuPct,
		MemoryUtilization: memPct,
		MemoryUsedBytes:   float64(usedKB) * 1024,
		MemoryTotalBytes:  float64(memTotalKB) * 1024,
	}, nil
}

func vmCPUUtilization(ctx context.Context, interval time.Duration) (float64, error) {
	initialSample, err := readCPUSample()
	if err != nil {
		return 0, err
	}

	if interval > 0 {
		timer := time.NewTimer(interval)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timer.C:
		}
	}

	finalSample, err := readCPUSample()
	if err != nil {
		return 0, err
	}

	if finalSample.total < initialSample.total || finalSample.idle < initialSample.idle {
		return 0, fmt.Errorf("invalid cpu sample delta: counter reset detected")
	}

	totalDelta := finalSample.total - initialSample.total
	idleDelta := finalSample.idle - initialSample.idle
	if totalDelta == 0 {
		return 0, fmt.Errorf("invalid cpu sample delta: total is 0")
	}

	usage := (float64(totalDelta-idleDelta) / float64(totalDelta)) * 100
	if usage < 0 {
		return 0, nil
	}

	if usage > 100 {
		usage = 100
	}

	return usage, nil
}

func readCPUSample() (cpuSample, error) {
	f, err := os.Open(procStatPath)
	if err != nil {
		return cpuSample{}, fmt.Errorf("unable to open %s: %w", procStatPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return cpuSample{}, fmt.Errorf("unable to read %s: %w", procStatPath, scanner.Err())
		}

		return cpuSample{}, fmt.Errorf("unable to read first line of %s", procStatPath)
	}

	return parseCPUSampleLine(scanner.Text())
}

func parseCPUSampleLine(line string) (cpuSample, error) {
	fields := strings.Fields(line)
	if len(fields) < 6 || fields[0] != "cpu" {
		return cpuSample{}, fmt.Errorf("invalid cpu stat line: %q", line)
	}

	values := make([]uint64, 0, len(fields)-1)
	for _, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuSample{}, fmt.Errorf("invalid cpu stat value %q: %w", field, err)
		}
		values = append(values, value)
	}

	total := uint64(0)
	for _, v := range values {
		total += v
	}

	if len(values) < 5 {
		return cpuSample{}, fmt.Errorf("invalid cpu stat values: %q", line)
	}

	idle := values[3] + values[4]
	return cpuSample{total: total, idle: idle}, nil
}

func vmMemoryInfoKB() (uint64, uint64, error) {
	b, err := os.ReadFile(procMemPath)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to read %s: %w", procMemPath, err)
	}

	return parseMemInfo(string(b))
}

func parseMemInfo(content string) (uint64, uint64, error) {
	var memTotalKB uint64
	var memAvailableKB uint64
	var memFreeKB uint64
	var buffersKB uint64
	var cachedKB uint64

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotalKB = value
		case "MemAvailable:":
			memAvailableKB = value
		case "MemFree:":
			memFreeKB = value
		case "Buffers:":
			buffersKB = value
		case "Cached:":
			cachedKB = value
		}
	}

	if scanner.Err() != nil {
		return 0, 0, fmt.Errorf("unable to parse %s: %w", procMemPath, scanner.Err())
	}

	if memTotalKB == 0 {
		return 0, 0, fmt.Errorf("missing MemTotal in %s", procMemPath)
	}

	if memAvailableKB == 0 {
		memAvailableKB = memFreeKB + buffersKB + cachedKB
		if memAvailableKB == 0 {
			return 0, 0, fmt.Errorf("missing MemAvailable in %s", procMemPath)
		}
	}

	if memAvailableKB > memTotalKB {
		memAvailableKB = memTotalKB
	}

	return memTotalKB, memAvailableKB, nil
}

func collectDockerResourceStats(ctx context.Context) ([]dockerResourceStat, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, dockerStatsCmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "docker", "stats", "--no-stream", "--format", dockerStatsFormat)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("docker command not found: %w", err)
		}

		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return nil, fmt.Errorf("docker stats command failed: %w", err)
		}

		return nil, fmt.Errorf("docker stats command failed: %w: %s", err, trimmed)
	}

	return parseDockerStatsOutput(string(out))
}

func parseDockerStatsOutput(output string) ([]dockerResourceStat, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []dockerResourceStat{}, nil
	}

	lines := strings.Split(output, "\n")
	stats := make([]dockerResourceStat, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			return nil, fmt.Errorf("unexpected docker stats output line: %q", line)
		}

		cpuPct, err := parsePercent(fields[2])
		if err != nil {
			return nil, fmt.Errorf("unable to parse docker cpu percentage %q: %w", fields[2], err)
		}

		memPct, err := parsePercent(fields[3])
		if err != nil {
			return nil, fmt.Errorf("unable to parse docker memory percentage %q: %w", fields[3], err)
		}

		memUsed, memLimit, err := parseDockerMemoryUsage(fields[4])
		if err != nil {
			return nil, fmt.Errorf("unable to parse docker memory usage %q: %w", fields[4], err)
		}

		stats = append(stats, dockerResourceStat{
			ContainerID:       strings.TrimSpace(fields[0]),
			ContainerName:     strings.TrimSpace(fields[1]),
			CPUUtilizationPct: cpuPct,
			MemoryUtilization: memPct,
			MemoryUsedBytes:   memUsed,
			MemoryLimitBytes:  memLimit,
		})
	}

	return stats, nil
}

func parsePercent(raw string) (float64, error) {
	raw = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(raw), "%"))
	if raw == "" || raw == "--" {
		return 0, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func parseDockerMemoryUsage(raw string) (float64, float64, error) {
	parts := strings.Split(raw, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid docker memory usage value %q", raw)
	}

	used, err := parseByteSize(parts[0])
	if err != nil {
		return 0, 0, err
	}

	limit, err := parseByteSize(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return used, limit, nil
}

func parseByteSize(raw string) (float64, error) {
	value := strings.TrimSpace(raw)
	if value == "" || value == "--" {
		return 0, nil
	}

	value = strings.ReplaceAll(value, " ", "")
	if value == "" {
		return 0, nil
	}

	numericPart := value
	unitPart := ""
	for idx, r := range value {
		if (r < '0' || r > '9') && r != '.' {
			numericPart = value[:idx]
			unitPart = strings.ToLower(strings.TrimSpace(value[idx:]))
			break
		}
	}

	number, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size value %q: %w", raw, err)
	}

	multiplier, ok := byteSizeMultipliers[unitPart]
	if !ok {
		return 0, fmt.Errorf("unknown byte size unit %q", unitPart)
	}

	return number * multiplier, nil
}

var byteSizeMultipliers = map[string]float64{
	"":    1,
	"b":   1,
	"k":   1e3,
	"kb":  1e3,
	"m":   1e6,
	"mb":  1e6,
	"g":   1e9,
	"gb":  1e9,
	"t":   1e12,
	"tb":  1e12,
	"ki":  1024,
	"kib": 1024,
	"mi":  1024 * 1024,
	"mib": 1024 * 1024,
	"gi":  1024 * 1024 * 1024,
	"gib": 1024 * 1024 * 1024,
	"ti":  1024 * 1024 * 1024 * 1024,
	"tib": 1024 * 1024 * 1024 * 1024,
}
