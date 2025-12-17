package monitor

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"website_change_monitor/internal/notifications"

	"github.com/PuerkitoBio/goquery"
	"github.com/sergi/go-diff/diffmatchpatch"
)

/* =======================
   STRUCTURI
======================= */

type ChangeRecord struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Diff      string    `json:"diff,omitempty"`
	Text      string    `json:"text,omitempty"`
}

type MonitorItem struct {
	URL       string         `json:"url"`
	Selector  string         `json:"selector"`
	Frequency int            `json:"frequency"` // MINUTE
	LastHash  string         `json:"last_hash"`
	UseJS     bool           `json:"use_js"` // folosește JS rendering
	History   []ChangeRecord `json:"history"`
}

/* =======================
   VARIABILE GLOBALE
======================= */

var (
	itemsInMemory []MonitorItem
	memMu         sync.Mutex
	started       = make(map[string]bool)
)

/* =======================
   LOAD / SAVE
======================= */

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

func SaveMonitorsFromMemory() {
	memMu.Lock()
	defer memMu.Unlock()
	_ = SaveMonitors(itemsInMemory)
}

/* =======================
   UTILS
======================= */

func HashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)
}

/* =======================
   CHECK URL
======================= */

func CheckURL(item *MonitorItem) (bool, string, string, error) {
	client := http.Client{Timeout: 20 * time.Second}

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

	var selectedText string

	if item.UseJS {
		// JS rendering pentru site-uri dinamice
		selectedText, err = FetchTextWithJS(item.URL, item.Selector)
		if err != nil {
			return false, "", "", err
		}
	} else {
		// site static
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return false, "", "", err
		}
		selectedText = strings.TrimSpace(doc.Find(item.Selector).Text())
	}

	newHash := HashContent([]byte(selectedText))

	// 🔹 PRIMA rulare → salvăm hash
	if item.LastHash == "" {
		item.LastHash = newHash
		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Text:      selectedText,
		})
		return false, "", selectedText, nil
	}

	// 🔹 SCHIMBARE
	if newHash != item.LastHash {
		oldText := ""
		if len(item.History) > 0 {
			oldText = item.History[len(item.History)-1].Text
		}

		dmp := diffmatchpatch.New()
		diff := dmp.DiffPrettyText(
			dmp.DiffMain(oldText, selectedText, false),
		)

		item.LastHash = newHash
		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Text:      selectedText,
			Diff:      diff,
		})

		return true, oldText, selectedText, nil
	}

	return false, "", selectedText, nil
}

/* =======================
   RUN MONITOR (SERVICE)
======================= */

func RunMonitorContinuously() {
	logFile, err := os.OpenFile("monitor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Cannot open log file:", err)
		return
	}
	log.SetOutput(logFile)

	log.Println("Starting monitoring service...")

	// Load initial
	itemsInMemory, err = LoadMonitors()
	if err != nil {
		log.Println("Failed to load monitors:", err)
		return
	}

	// Start monitors pentru toate itemele inițiale
	memMu.Lock()
	for i := range itemsInMemory {
		startMonitor(&itemsInMemory[i])
	}
	memMu.Unlock()

	reloadTicker := time.NewTicker(15 * time.Second)
	defer reloadTicker.Stop()

	for range reloadTicker.C {
		// Reload JSON pentru a detecta iteme noi
		newItems, err := LoadMonitors()
		if err != nil {
			log.Println("Failed to reload monitors:", err)
			continue
		}

		memMu.Lock()
		itemsInMemory = newItems
		// Pornim monitor pentru orice URL nou
		for i := range itemsInMemory {
			startMonitor(&itemsInMemory[i])
		}
		memMu.Unlock()
	}
}

/* =======================
   START SINGLE MONITOR
======================= */

func startMonitor(item *MonitorItem) {
	key := item.URL + "|" + item.Selector
	if started[key] {
		return
	}
	started[key] = true

	log.Println("Starting monitor for", item.URL)

	go func(it *MonitorItem) {
		// ✅ CHECK IMEDIAT → hash inițial
		_, _, _, err := CheckURL(it)
		if err != nil {
			log.Println("Initial check failed for", it.URL, err)
		} else {
			SaveMonitorsFromMemory()
			log.Println("Initial hash saved for", it.URL)
		}

		ticker := time.NewTicker(time.Duration(it.Frequency) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			changed, oldText, newText, err := CheckURL(it)
			if err != nil {
				log.Println("Error checking", it.URL, err)
				continue
			}

			if changed {
				log.Println("CHANGED:", it.URL)
				SaveMonitorsFromMemory()

				msg := fmt.Sprintf(
					"Old value: %s\nNew value: %s",
					oldText, newText,
				)

				if err := notifications.SendEmailNotification(it.URL, msg); err != nil {
					log.Println("Email error:", err)
				}
			}
		}
	}(item)
}
