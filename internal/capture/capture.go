package capture

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bjluckow/harchiver/internal/cdp"
	hartype "github.com/bjluckow/harchiver/pkg/har-type"
	harutil "github.com/bjluckow/harchiver/pkg/har-util"
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

	// Enable network tracking
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return fmt.Errorf("enable network: %w", err)
	}

	rec := cdp.NewRecorder()
	rec.Listen(ctx)

	for _, u := range opts.URLs {
		if err := chromedp.Run(ctx,
			chromedp.Navigate(u),
			chromedp.WaitReady("body", chromedp.ByQuery),
		); err != nil {
			return fmt.Errorf("navigate %s: %w", u, err)
		}
	}

	archive := &hartype.HttpArchive{
		Log: hartype.Log{
			Entries: rec.Entries(),
		},
	}

	if err := harutil.Write(archive, opts.Output); err != nil {
		return fmt.Errorf("write HAR: %w", err)
	}

	return nil
}
