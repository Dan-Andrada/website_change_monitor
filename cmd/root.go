package cmd

import (
	"fmt"
	"os"
	"strconv"
	"website_change_monitor/internal/monitor"
)

func Execute() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: monitor <command>")
		return
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 5 {
			fmt.Println("Usage: monitor add <url> <css-selector> <frequency>")
			return
		}

		url := os.Args[2]
		selector := os.Args[3]
		frequency, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Frequency must be a number")
			return
		}

		err = monitor.AddURL(url, selector, frequency)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("URL added:", url, "Selector:", selector, "Frequency:", frequency)

	case "check":
		items, err := monitor.LoadMonitors()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for i := range items {
			changed, oldVal, newVal, err := monitor.CheckURL(&items[i])
			if err != nil {
				fmt.Println("Error checking:", items[i].URL, err)
				continue
			}
			if changed {
				fmt.Println("🔴 CHANGE DETECTED:", items[i].URL)
				fmt.Println("   Old:", oldVal)
				fmt.Println("   New:", newVal)
			} else {
				fmt.Println("🟢 NO CHANGE:", items[i].URL)
			}

		}

		monitor.SaveMonitors(items)

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
