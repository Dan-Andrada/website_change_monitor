package main

import (
	"fmt"
	"os"
	"website_change_monitor/internal/monitor"
)

func main() {

	if len(os.Args) > 1 && os.Args[1] == "add" && len(os.Args) > 2 {
		url := os.Args[2]
		err := monitor.AddURL(url)
		if err != nil {
			fmt.Println("Eroare la adăugarea URL:", err)
		} else {
			fmt.Println("URL adăugat cu succes:", url)
		}
		return
	}

	changed, err := monitor.CheckAll()
	if err != nil {
		fmt.Println("Eroare la verificarea site-urilor:", err)
		return
	}

	if len(changed) == 0 {
		fmt.Println("Nu s-au detectat schimbări.")
	} else {
		fmt.Println("Site-uri modificate:")
		for _, item := range changed {
			fmt.Println("-", item.URL)
		}
	}
}
