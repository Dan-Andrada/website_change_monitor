package monitor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ParsePrice transformă un string de tip "548,06 Lei" în float64
func ParsePrice(s string) (float64, error) {
	s = strings.ReplaceAll(s, "Lei", "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "") // elimină separatorul de mii
	s = strings.ReplaceAll(s, ",", ".")
	return strconv.ParseFloat(s, 64)
}

// ChangeRecord păstrează istoricul modificărilor
type ChangeRecord struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Diff      string    `json:"diff,omitempty"`
	Text      string    `json:"text,omitempty"`
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

// CheckURL verifică un URL și returnează schimbarea, plus vechea și noua valoare
func CheckURL(item *MonitorItem) (changed bool, oldValue, newValue string, err error) {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", item.URL, nil)
	if err != nil {
		return false, "", "", err
	}
	req.Header.Set("User-Agent", "Go-Site-Monitor/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return false, "", "", err
	}

	selectedText := doc.Find(item.Selector).Text()
	selectedText = strings.TrimSpace(selectedText)
	newHash := HashContent([]byte(selectedText))

	// Prima rulare
	if item.LastHash == "" {
		item.LastHash = newHash
		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Text:      selectedText,
		})
		return false, "", selectedText, nil
	}

	// Dacă s-a schimbat
	if newHash != item.LastHash {
		oldText := ""
		if len(item.History) > 0 {
			oldText = item.History[len(item.History)-1].Text
		}

		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(oldText, selectedText, false)
		diffText := dmp.DiffPrettyText(diffs)

		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Text:      selectedText,
			Diff:      diffText,
		})
		item.LastHash = newHash

		return true, oldText, selectedText, nil
	}

	return false, "", selectedText, nil
}

// CheckAll verifică toate URL-urile și afișează schimbările cu diferența de preț
func CheckAll() ([]MonitorItem, error) {
	items, err := LoadMonitors()
	if err != nil {
		return nil, err
	}

	changed := []MonitorItem{}

	for i := range items {
		changedNow, oldText, newText, err := CheckURL(&items[i])
		if err != nil {
			fmt.Println("Eroare la verificare:", items[i].URL, err)
			continue
		}
		if changedNow {
			fmt.Println("🔴 CHANGED:", items[i].URL)
			fmt.Println("Old value:", oldText)
			fmt.Println("New value:", newText)

			oldPrice, err1 := ParsePrice(oldText)
			newPrice, err2 := ParsePrice(newText)
			if err1 == nil && err2 == nil {
				diff := newPrice - oldPrice
				if diff > 0 {
					fmt.Printf("Price increased by %.2f Lei\n", diff)
				} else if diff < 0 {
					fmt.Printf("Price decreased by %.2f Lei\n", -diff)
				} else {
					fmt.Println("Price unchanged")
				}
			}
			changed = append(changed, items[i])
		} else {
			fmt.Println("🟢 NO CHANGE:", items[i].URL)
		}
	}

	err = SaveMonitors(items)
	if err != nil {
		return nil, err
	}

	return changed, nil
}
