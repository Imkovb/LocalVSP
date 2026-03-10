package docker

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func GetSystemInfo() (SystemInfo, error) {
	info := SystemInfo{}

	versionOut, err := runDockerCmd(defaultTimeout, "", "version", "--format", "{{.Server.Version}}")
	if err == nil {
		info.DockerVersion = strings.TrimSpace(string(versionOut))
	}

	psOut, err := runDockerCmd(defaultTimeout, "", "ps", "-a", "--format", "{{.Status}}")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(psOut)))
		for scanner.Scan() {
			line := scanner.Text()
			info.TotalContainers++
			if strings.HasPrefix(strings.ToLower(line), "up") {
				info.RunningContainers++
			}
		}
	}

	info.DiskUsage = getDiskUsage()
	info.LoadAvg = getSystemLoad()
	info.MemoryUsage = getMemoryUsage()
	return info, nil
}

func getSystemLoad() string {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "N/A"
	}
	parts := strings.Fields(string(data))
	if len(parts) < 1 {
		return "N/A"
	}
	load1, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return "N/A"
	}

	numCPU := getCPUCount()
	if numCPU <= 0 {
		numCPU = 1
	}
	pct := (load1 / float64(numCPU)) * 100
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%.0f%%", pct)
}

func getCPUCount() int {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 1
	}
	count := 0
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "processor") {
			count++
		}
	}
	return count
}

func getMemoryUsage() string {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "N/A"
	}
	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				memAvailable, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}
	if memTotal > 0 {
		used := memTotal - memAvailable
		percent := float64(used) / float64(memTotal) * 100
		return fmt.Sprintf("%.1f%% (%.1f/%.1f GB)", percent, float64(used)/1024/1024, float64(memTotal)/1024/1024)
	}
	return "N/A"
}

func getDiskUsage() string {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/opt/localvsp", &stat); err != nil {
		return "N/A"
	}
	bsize := uint64(stat.Bsize)
	total := float64(stat.Blocks*bsize) / 1024 / 1024 / 1024
	free := float64(stat.Bavail*bsize) / 1024 / 1024 / 1024
	if total == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.1f GB free / %.1f GB total", free, total)
}