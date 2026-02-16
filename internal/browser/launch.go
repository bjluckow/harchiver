package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

type LaunchOptions struct {
	ExecPath string
	Headless bool
	Stealth  bool
}

// Launch starts a local Chrome instance
func Launch(parent context.Context, opts *LaunchOptions) (*Context, error) {
	cdpOpts := chromedp.DefaultExecAllocatorOptions[:]

	if opts.ExecPath != "" {
		cdpOpts = append(cdpOpts, chromedp.ExecPath(opts.ExecPath))
	}

	if !opts.Headless {
		cdpOpts = append(cdpOpts, chromedp.Flag("headless", false))
	}

	if opts.Stealth {
		cdpOpts = append(cdpOpts,
			chromedp.Flag("enable-automation", false),
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		)
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, cdpOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	return &Context{
		Ctx: ctx,
		Cancel: func() {
			cancel()
			allocCancel()
		},
	}, nil
}
