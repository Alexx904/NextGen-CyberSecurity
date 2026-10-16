package service

import (
	"math"
	"nextgen-cybersec/internal/scanner"
	"strings"
)

// estimateScanProgress maps task-local Nmap telemetry to a coarse total estimate.
// Windows are heuristics, not measurements of total work; only terminal status proves completion.
func estimateScanProgress(progress scanner.TaskProgress, elapsed, previous float64) (phase string, total float64, remaining *float64) {
	task := strings.ToLower(progress.Task)
	start, end := 0.0, 0.0
	switch {
	case strings.Contains(task, "ping"), strings.Contains(task, "arp"), strings.Contains(task, "dns"), strings.Contains(task, "resolve"):
		phase, start, end = "HOST_DISCOVERY", 0, 5
	case strings.Contains(task, "nse"), strings.Contains(task, "script"):
		phase, start, end = "NSE", 80, 90
	case strings.Contains(task, "service"), strings.Contains(task, "version"), strings.Contains(task, "os detection"):
		phase, start, end = "SERVICE_DETECTION", 65, 80
	case strings.Contains(task, "scan"):
		phase, start, end = "PORT_SCAN", 5, 65
	default:
		return "", previous, nil
	}
	if progress.Percent == nil {
		return "", previous, nil
	}
	total = math.Max(previous, math.Min(99, start+(end-start)*(*progress.Percent)/100))
	// Extrapolate the observed total rate; native remaining is a lower bound for current task.
	if phase != "HOST_DISCOVERY" && total >= 10 && elapsed >= 5 {
		seconds := elapsed * (100 - total) / total
		if progress.Remaining != nil {
			seconds = math.Max(seconds, *progress.Remaining)
		}
		if seconds > 0 && !math.IsInf(seconds, 0) && !math.IsNaN(seconds) {
			remaining = &seconds
		}
	}
	return
}
