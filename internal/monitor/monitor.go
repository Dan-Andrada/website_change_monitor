package monitor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"website_change_monitor/internal/notifications"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/sergi/go-diff/diffmatchpatch"
)

/* =======================
   STRUCTURI
======================= */

type ChangeRecord struct {
	Hash       string    `json:"hash"`
	Timestamp  time.Time `json:"timestamp"`
	Text       string    `json:"text"`
	Diff       string    `json:"diff,omitempty"`
	Screenshot string    `json:"screenshot,omitempty"`
}

type MonitorItem struct {
	URL       string         `json:"url"`
	Selector  string         `json:"selector"`
	Frequency int            `json:"frequency"` // minute
	LastHash  string         `json:"last_hash"`
	History   []ChangeRecord `json:"history"`
}

/* =======================
   VARIABILE
======================= */

var (
	monitors []MonitorItem
	mu       sync.Mutex
	started  = map[string]bool{}
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
	return items, json.Unmarshal(data, &items)
}

func SaveMonitors() {
	mu.Lock()
	defer mu.Unlock()

	data, _ := json.MarshalIndent(monitors, "", "  ")
	_ = os.WriteFile("data/monitors.json", data, 0644)
}

/* =======================
   UTILS
======================= */

func hashText(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

/* =======================
   SCREENSHOT
======================= */

func captureScreenshot(url string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(`C:\Program Files\Google\Chrome\Application\chrome.exe`),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var buf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return "", err
	}

	if len(buf) == 0 {
		return "", fmt.Errorf("empty screenshot buffer")
	}

	_ = os.MkdirAll("screenshots", 0755)
	filename := fmt.Sprintf("screenshots/%d.png", time.Now().UnixNano())
	err = os.WriteFile(filename, buf, 0644)

	return filepath.ToSlash(filename), err
}

/* =======================
   CHECK URL
======================= */

func check(item *MonitorItem) (bool, string, string, error) {
	resp, err := http.Get(item.URL)
	if err != nil {
		return false, "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return false, "", "", err
	}

	text := strings.TrimSpace(doc.Find(item.Selector).Text())
	newHash := hashText(text)

	if item.LastHash == "" {
		item.LastHash = newHash
		item.History = append(item.History, ChangeRecord{
			Hash:      newHash,
			Timestamp: time.Now(),
			Text:      text,
		})
		return false, "", text, nil
	}

	if newHash != item.LastHash {
		old := item.History[len(item.History)-1].Text

		dmp := diffmatchpatch.New()
		diff := dmp.DiffPrettyText(dmp.DiffMain(old, text, false))

		log.Println("📸 About to capture screenshot")
		ss, err := captureScreenshot(item.URL)
		if err != nil {
			log.Println("❌ Screenshot error:", err)
		} else {
			log.Println("✅ Screenshot path:", ss)
		}

		item.LastHash = newHash
		item.History = append(item.History, ChangeRecord{
			Hash:       newHash,
			Timestamp:  time.Now(),
			Text:       text,
			Diff:       diff,
			Screenshot: ss,
		})

		return true, old, text, nil
	}

	return false, "", text, nil
}

/* =======================
   MONITOR LOOP
======================= */

func start(item *MonitorItem) {
	key := item.URL + "|" + item.Selector
	if started[key] {
		return
	}
	started[key] = true

	go func() {
		_, _, _, _ = check(item)
		SaveMonitors()

		ticker := time.NewTicker(time.Duration(item.Frequency) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			changed, oldV, newV, err := check(item)
			if err != nil {
				log.Println("Check error:", err)
				continue
			}

			if changed {
				SaveMonitors()
				msg := fmt.Sprintf(
					"Change detected!\n\nOld:\n%s\n\nNew:\n%s",
					oldV, newV,
				)
				_ = notifications.SendEmailNotification(item.URL, msg)
				log.Println("CHANGED:", item.URL)
			}
		}
	}()
}

/* =======================
   RUN
======================= */

func RunMonitorContinuously() {
	logFile, _ := os.OpenFile("monitor.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	log.SetOutput(logFile)

	var err error
	monitors, err = LoadMonitors()
	if err != nil {
		log.Fatal(err)
	}

	for i := range monitors {
		start(&monitors[i])
	}

	select {}
}
