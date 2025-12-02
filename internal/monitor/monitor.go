package monitor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ChangeRecord păstrează istoricul modificărilor
type ChangeRecord struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Diff      string    `json:"diff,omitempty"`
	Text      string    `json:"text,omitempty"` // textul complet pentru diff
}

// MonitorItem reprezintă un site monitorizat
type MonitorItem struct {
	URL       string         `json:"url"`
	Selector  string         `json:"selector"`
	Frequency int            `json:"frequency"`
	LastHash  string         `json:"last_hash"`
	History   []ChangeRecord `json:"history"`
}

// LoadMonitors încarcă lista de site-uri din JSON
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

// SaveMonitors salvează lista de site-uri în JSON
func SaveMonitors(items []MonitorItem) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("data/monitors.json", data, 0644)
}

// AddURL adaugă un URL nou în monitorizare
func AddURL(url, selector string, frequency int) error {
	items, err := LoadMonitors()
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.URL == url && item.Selector == selector {
			return fmt.Errorf("URL already exists with this selector")
		}
	}

	items = append(items, MonitorItem{
		URL:       url,
		Selector:  selector,
		Frequency: frequency,
	})
	return SaveMonitors(items)
}

// HashContent calculează SHA256 pentru un text
func HashContent(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

// CheckURL verifică un URL și actualizează istoricul dacă s-a schimbat
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

	// parsează HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return false, err
	}

	selectedText := doc.Find(item.Selector).Text()
	newHash := HashContent([]byte(selectedText))

	if item.LastHash == "" {
		item.LastHash = newHash
		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Diff:      "",
			Text:      selectedText,
		})
		return false, nil
	}

	if newHash != item.LastHash {
		// generează diff
		dmp := diffmatchpatch.New()
		oldText := ""
		if len(item.History) > 0 {
			oldText = item.History[len(item.History)-1].Text
		}
		diffs := dmp.DiffMain(oldText, selectedText, false)
		diffText := dmp.DiffPrettyText(diffs)

		// salvează în istoric
		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Diff:      diffText,
			Text:      selectedText,
		})

		item.LastHash = newHash
		return true, nil
	}

	return false, nil
}

// CheckAll verifică toate URL-urile și salvează hash-urile actualizate
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

	err = SaveMonitors(items)
	if err != nil {
		return nil, err
	}

	return changed, nil
}
