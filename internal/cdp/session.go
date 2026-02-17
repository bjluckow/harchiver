package cdp

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/chromedp/cdproto/har"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

type Session struct {
	mu        sync.Mutex
	recorder  *Recorder
	targets   map[target.ID]context.CancelFunc
	parentCtx context.Context
}

func NewSession(ctx context.Context) *Session {
	return &Session{
		recorder:  NewRecorder(),
		targets:   make(map[target.ID]context.CancelFunc),
		parentCtx: ctx,
	}
}

func (s *Session) Start() error {
	// Tell browser to pause new targets so we don't miss requests while we attach
	if err := chromedp.Run(s.parentCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			return target.SetAutoAttach(true, true).WithFlatten(true).Do(ctx)
		}),
	); err != nil {
		return fmt.Errorf("set auto attach: %w", err)
	}

	targets, err := chromedp.Targets(s.parentCtx)
	if err != nil {
		return err
	}
	for _, t := range targets {
		if t.Type == "page" {
			s.attachTarget(t.TargetID)
		}
	}

	chromedp.ListenBrowser(s.parentCtx, func(ev any) {
		switch e := ev.(type) {
		case *target.EventTargetCreated:
			if e.TargetInfo.Type == "page" {
				go s.attachTarget(e.TargetInfo.TargetID)
			}
		case *target.EventTargetDestroyed:
			s.detachTarget(e.TargetID)
		}
	})

	return nil
}

func (s *Session) attachTarget(id target.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.targets[id]; ok {
		return // already attached
	}

	tabCtx, cancel := chromedp.NewContext(s.parentCtx,
		chromedp.WithTargetID(id),
	)

	s.recorder.ListenTarget(tabCtx, string(id))

	if err := chromedp.Run(tabCtx, network.Enable()); err != nil {
		log.Printf("enable network on %s: %v", id, err)
		cancel()
		return
	}

	s.targets[id] = cancel
}

func (s *Session) detachTarget(id target.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, ok := s.targets[id]; ok {
		cancel()
		delete(s.targets, id)
	}
}

func (s *Session) HAR() *har.HAR {
	return &har.HAR{
		Log: &har.Log{
			Version: "1.2",
			Creator: &har.Creator{
				Name:    "harchiver",
				Version: "1.1.0",
			},
			Entries: s.recorder.entries,
		},
	}
}

func (s *Session) Navigate(ctx context.Context, urls []string) error {
	for _, u := range urls {
		if err := chromedp.Run(ctx,
			chromedp.Navigate(u), chromedp.WaitReady("body", chromedp.ByQuery)); err != nil {
			return fmt.Errorf("navigate %s: %w", u, err)
		}
	}
	return nil
}
