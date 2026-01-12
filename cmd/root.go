package cmd

import (
	"fmt"
	"os"
	"strconv"

	"website_change_monitor/internal/monitor"
)

func Execute() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {

	case "add":
		if len(os.Args) != 5 {
			fmt.Println("Usage:")
			fmt.Println("  gomonitor add <url> <selector> <frequency_minutes>")
			return
		}

		freq, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Frequency must be a number (minutes)")
			return
		}

		err = monitor.Add(
			os.Args[2],
			os.Args[3],
			freq,
		)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("✅ Monitor added")

	case "run":
		monitor.RunMonitorContinuously()

	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  gomonitor add <url> <selector> <frequency_minutes>")
	fmt.Println("  gomonitor run")
}
