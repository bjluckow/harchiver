package cdp

import (
	"context"
	"fmt"
	"sync"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

type Client struct {
	browserCtx context.Context
	contextID  cdp.BrowserContextID

	mu        sync.Mutex
	tabs      map[string]*Tab
	activeTab *Tab
	nextIndex int
}

type Tab struct {
	ID     target.ID
	Ctx    context.Context
	Cancel context.CancelFunc
	Title  string
	URL    string
}

func NewClient(browserCtx context.Context) *Client {
	return &Client{
		browserCtx: browserCtx,
		tabs:       make(map[string]*Tab),
	}
}

// NewContext creates an isolated browser context (like incognito).
// If not called, the client operates in the default context.
func (c *Client) NewContext() error {
	id, err := target.CreateBrowserContext().Do(cdp.WithExecutor(c.browserCtx,
		chromedp.FromContext(c.browserCtx).Browser))
	if err != nil {
		return fmt.Errorf("create browser context: %w", err)
	}
	c.contextID = id
	return nil
}

func (c *Client) NewTab(url string) (*Tab, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	targetID, err := target.CreateTarget(url).
		WithBrowserContextID(c.contextID).
		Do(cdp.WithExecutor(c.browserCtx,
			chromedp.FromContext(c.browserCtx).Browser))
	if err != nil {
		return nil, fmt.Errorf("create target: %w", err)
	}

	tabCtx, cancel := chromedp.NewContext(c.browserCtx,
		chromedp.WithTargetID(targetID),
	)

	key := fmt.Sprintf("%d", c.nextIndex)
	c.nextIndex++

	tab := &Tab{
		ID:     targetID,
		Ctx:    tabCtx,
		Cancel: cancel,
		URL:    url,
	}

	c.tabs[key] = tab
	c.activeTab = tab
	return tab, nil
}

func (c *Client) Navigate(url string) error {
	if c.activeTab == nil {
		_, err := c.NewTab(url)
		return err
	}
	c.activeTab.URL = url
	return chromedp.Run(c.activeTab.Ctx, chromedp.Navigate(url))
}

func (c *Client) SwitchTab(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tab, ok := c.tabs[key]
	if !ok {
		return fmt.Errorf("no tab %q", key)
	}
	c.activeTab = tab
	return nil
}

func (c *Client) CloseTab(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	tab, ok := c.tabs[key]
	if !ok {
		return fmt.Errorf("no tab %q", key)
	}

	tab.Cancel()
	delete(c.tabs, key)

	if c.activeTab == tab {
		c.activeTab = nil
	}
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tab := range c.tabs {
		tab.Cancel()
	}
	c.tabs = make(map[string]*Tab)
	c.activeTab = nil

	if c.contextID != "" {
		return target.DisposeBrowserContext(c.contextID).
			Do(cdp.WithExecutor(c.browserCtx,
				chromedp.FromContext(c.browserCtx).Browser))
	}
	return nil
}
