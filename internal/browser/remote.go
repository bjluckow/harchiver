package browser

import (
	"context"

	"github.com/chromedp/chromedp"
)

func Remote(parent context.Context, endpoint string) (*Context, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(parent, endpoint)

	ctx, cancel := chromedp.NewContext(allocCtx)

	return &Context{
		Ctx: ctx,
		Cancel: func() {
			cancel()
			allocCancel()
		},
	}, nil
}
