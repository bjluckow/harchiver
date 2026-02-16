package capture

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bjluckow/harchiver/internal/cdp"
	"github.com/bjluckow/harchiver/pkg/har"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Options struct {
	URLs    []string
	Output  io.Writer
	Timeout time.Duration
}

func Run(ctx context.Context, opts Options) error {
	if len(opts.URLs) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	rec := cdp.NewRecorder()
	rec.Listen(ctx)

	// Enable network tracking
	// MUST occur between rec.Listen() and any navigations
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return fmt.Errorf("enable network: %w", err)
	}

	for _, u := range opts.URLs {
		if err := chromedp.Run(ctx,
			chromedp.Navigate(u),
			chromedp.WaitReady("body", chromedp.ByQuery),
		); err != nil {
			return fmt.Errorf("navigate %s: %w", u, err)
		}
	}

	archive := &har.HttpArchive{
		Log: har.Log{
			Entries: rec.Entries(),
		},
	}

	if err := har.Write(archive, opts.Output); err != nil {
		return fmt.Errorf("write HAR: %w", err)
	}

	return nil
}
