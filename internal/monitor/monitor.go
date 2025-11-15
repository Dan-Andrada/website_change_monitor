package monitor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type MonitorItem struct {
	URL      string `json:"url"`
	LastHash string `json:"last_hash"`
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

	for _, item := range items {
		if item.URL == url {
			return fmt.Errorf("URL already exists")
		}
	}

	items = append(items, MonitorItem{URL: url})
	return SaveMonitors(items)
}

func HashContent(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// vedem daca s a schimbat URL
func CheckURL(item *MonitorItem) (bool, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", item.URL, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", "Go-Site-Monitor/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	newHash := HashContent(body)

	if item.LastHash == "" {
		item.LastHash = newHash
		return false, nil // prima rulare
	}

	if newHash != item.LastHash {
		item.LastHash = newHash
		return true, nil
	}

	return false, nil
}

// verifica toate URL-urile si salveaza hash urile actualizate
func CheckAll() ([]MonitorItem, error) {
	items, err := LoadMonitors()
	if err != nil {
		return nil, err
	}

	changed := []MonitorItem{}

	for i := range items {
		isChanged, err := CheckURL(&items[i])
		if err != nil {
			fmt.Println("Eroare la verificarea:", items[i].URL, err)
			continue
		}
		if isChanged {
			changed = append(changed, items[i])
		}
	}

	// salvez toate hash-urile actualizate
	err = SaveMonitors(items)
	if err != nil {
		return nil, err
	}

	return changed, nil
}
