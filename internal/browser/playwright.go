package browser

import (
	"fmt"
	"time"

	pw "github.com/playwright-community/playwright-go"
)

type FetchOptions struct {
	URLs     []string
	HARPath  string
	Timeout  time.Duration
	Headless bool
}

// FetchWithHAR visits all URLs in one browser context and records a HAR.
func FetchWithHAR(opts FetchOptions) error {
	if len(opts.URLs) == 0 {
		return fmt.Errorf("no URLs provided")
	}
	if opts.HARPath == "" {
		return fmt.Errorf("HARPath is required")
	}

	if err := pw.Install(); err != nil {
		// ignore if already installed
	}

	pwInstance, err := pw.Run()
	if err != nil {
		return fmt.Errorf("playwright run: %w", err)
	}
	defer pwInstance.Stop()

	browser, err := pwInstance.Chromium.Launch(pw.BrowserTypeLaunchOptions{
		Headless: pw.Bool(opts.Headless),
	})
	if err != nil {
		return fmt.Errorf("launch chromium: %w", err)
	}
	defer browser.Close()

	context, err := browser.NewContext(pw.BrowserNewContextOptions{
		RecordHarPath: pw.String(opts.HARPath),
	})
	if err != nil {
		return fmt.Errorf("new context: %w", err)
	}

	for _, u := range opts.URLs {
		page, err := context.NewPage()
		if err != nil {
			return fmt.Errorf("new page: %w", err)
		}

		_, err = page.Goto(u, pw.PageGotoOptions{
			WaitUntil: pw.WaitUntilStateNetworkidle,
			Timeout:   pw.Float(float64(opts.Timeout.Milliseconds())),
		})
		if err != nil {
			_ = page.Close()
			return fmt.Errorf("goto %s: %w", u, err)
		}

		_ = page.Close()
	}

	// HAR is finalized here
	if err := context.Close(); err != nil {
		return fmt.Errorf("close context: %w", err)
	}

	return nil
}
