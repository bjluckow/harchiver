package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

type LaunchOptions struct {
	ExecPath string
	Headless bool
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
