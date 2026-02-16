package browser

import "context"

// Context creates a chromedp context and returns it with a cancel function
// How the browser is started depends on implementation (i.e., launch vs remote)
type Context struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}
