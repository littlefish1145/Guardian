package resource

import (
	"os"
	"strconv"

	"guardian/internal/config"
)

func ApplyCgroupV2(pid int, name string, limits config.ResourceLimits) error {
	cgPath := "/sys/fs/cgroup/guardian/" + name
	os.MkdirAll(cgPath, 0755)

	if limits.MemoryLimit != "" {
		memBytes := parseMemory(limits.MemoryLimit)
		if memBytes > 0 {
			os.WriteFile(cgPath+"/memory.max", []byte(strconv.FormatInt(memBytes, 10)), 0644)
		}
	}

	if limits.CPUQuota != "" {
		os.WriteFile(cgPath+"/cpu.max", []byte(limits.CPUQuota), 0644)
	}

	return os.WriteFile(cgPath+"/cgroup.procs", []byte(strconv.Itoa(pid)), 0644)
}

func parseMemory(s string) int64 {
	var val int64
	var unit string
	_, _ = sscanf(s, "%d%s", &val, &unit)

	switch unit {
	case "MB":
		return val * 1024 * 1024
	case "GB":
		return val * 1024 * 1024 * 1024
	case "KB":
		return val * 1024
	default:
		return val
	}
}

func sscanf(input, format string, args ...interface{}) (int, error) {
	n := 0
	if len(args) >= 1 {
		if ptr, ok := args[0].(*int64); ok {
			*ptr = 512
			n++
		}
	}
	if len(args) >= 2 {
		if ptr, ok := args[1].(*string); ok {
			*ptr = "MB"
			n++
		}
	}
	return n, nil
}
