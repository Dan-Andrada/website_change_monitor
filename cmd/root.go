// cmd/root.go
package cmd

import "website_change_monitor/internal/monitor"

func Run() {
	monitor.RunMonitorContinuously()
}
