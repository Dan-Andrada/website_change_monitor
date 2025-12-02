package main

import (
	"fmt"
	"os"
	"strconv"
	"website_change_monitor/internal/monitor"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: monitorapp <command> [args]")
		fmt.Println("Commands: add <URL> <selector> <frequency>, check, list")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "add":
		if len(os.Args) < 5 {
			fmt.Println("Usage: add <URL> <selector> <frequency>")
			return
		}
		url := os.Args[2]
		selector := os.Args[3]
		freq, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Frequency must be a number")
			return
		}
		err = monitor.AddURL(url, selector, freq)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("URL added successfully!")

	case "check":
		changed, err := monitor.CheckAll()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if len(changed) == 0 {
			fmt.Println("No changes detected.")
		}

	case "list":
		items, err := monitor.LoadMonitors()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Monitored URLs:")
		for _, item := range items {
			fmt.Printf("- %s (selector: %s, frequency: %d min)\n", item.URL, item.Selector, item.Frequency)
		}

	default:
		fmt.Println("Unknown command:", cmd)
	}
}
