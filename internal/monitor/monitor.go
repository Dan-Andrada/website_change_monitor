package monitor

import (
	"encoding/json"
	"os"
)

type MonitorItem struct {
	URL string `json:"url"`
}

func LoadMonitors() ([]MonitorItem, error) {
	data, err := os.ReadFile("data/monitors.json")
	if err != nil {
		if os.IsNotExist(err) {
			return []MonitorItem{}, nil
		}
		return nil, err
	}

	var items []MonitorItem
	err = json.Unmarshal(data, &items)
	return items, err
}

func SaveMonitors(items []MonitorItem) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/monitors.json", data, 0644)
}

func AddURL(url string) error {
	items, err := LoadMonitors()
	if err != nil {
		return err
	}

	items = append(items, MonitorItem{URL: url})
	return SaveMonitors(items)
}
