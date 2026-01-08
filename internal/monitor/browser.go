package monitor

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// FetchTextWithJS încarcă pagina cu JS și extrage textul selectorului
func FetchTextWithJS(url, selector string) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// timeout hard
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var text string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)

	return text, err
}
