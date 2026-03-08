package cdputil

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

func Connect(parent context.Context, endpoint string) (context.Context, context.CancelFunc, error) {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(parent, endpoint)
	ctx, cancel := chromedp.NewContext(allocCtx)

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		allocCancel()
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	return ctx, func() {
		cancel()
		allocCancel()
	}, nil
}
