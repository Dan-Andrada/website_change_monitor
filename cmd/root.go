package cmd

import (
	"fmt"
	"os"
	"website_change_monitor/internal/monitor"
)

func Execute() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: monitor <command>")
		return
	}

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: monitor add <url>")
			return
		}
		url := os.Args[2]
		err := monitor.AddURL(url)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("URL added:", url)
	case "check":
		items, err := monitor.LoadMonitors()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for i := range items {
			changed, err := monitor.CheckURL(&items[i])
			if err != nil {
				fmt.Println("Error checking:", items[i].URL, err)
				continue
			}

			if changed {
				fmt.Println("🔴 CHANGED:", items[i].URL)
			} else {
				fmt.Println("🟢 NO CHANGE:", items[i].URL)
			}
		}

		monitor.SaveMonitors(items)

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
