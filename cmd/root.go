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
	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}
