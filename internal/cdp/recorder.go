package cdp

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/har"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Recorder listens to CDP network events and builds HAR entries
type Recorder struct {
	mu       sync.Mutex
	requests map[network.RequestID]*har.Entry
	entries  []*har.Entry
	pages    map[string]*har.Page
}

func NewRecorder() *Recorder {
	return &Recorder{
		requests: make(map[network.RequestID]*har.Entry),
		pages:    make(map[string]*har.Page),
	}
}

func (r *Recorder) registerPage(pageID, title string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pages[pageID] = &har.Page{
		ID:              pageID,
		StartedDateTime: time.Now().UTC().Format(time.RFC3339Nano),
		Title:           title,
		PageTimings:     &har.PageTimings{},
	}
}

func (r *Recorder) ListenTarget(ctx context.Context, pageID string) {
	r.registerPage(pageID, pageID) // TODO: pass human-readable titles
	chromedp.ListenTarget(ctx, func(ev any) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			r.onRequest(e, pageID)
		case *network.EventResponseReceived:
			r.onResponse(e)
		case *network.EventLoadingFinished:
			r.onLoadingFinished(ctx, e)
		case *network.EventLoadingFailed:
			r.onLoadingFailed(e)
		}
	})
}

func (r *Recorder) onRequest(e *network.EventRequestWillBeSent, pageID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	headers := make([]*har.NameValuePair, 0, len(e.Request.Headers))
	for k, v := range e.Request.Headers {
		s, _ := v.(string)
		headers = append(headers, &har.NameValuePair{
			Name: k, Value: s,
		})
	}

	entry := &har.Entry{
		StartedDateTime: e.WallTime.Time().UTC().Format(time.RFC3339Nano),
		Request: &har.Request{
			Method:      e.Request.Method,
			URL:         e.Request.URL,
			HTTPVersion: "HTTP/1.1",
			Headers:     headers,
		},
		Response: &har.Response{
			Status: -1, // sentinel until we get the response
		},
	}

	if e.Request.HasPostData && len(e.Request.PostDataEntries) > 0 {
		var text strings.Builder
		for _, entry := range e.Request.PostDataEntries {
			text.WriteString(entry.Bytes)
		}

		mimeType, _ := headerValue(e.Request.Headers, "content-type")
		entry.Request.PostData = &har.PostData{
			MimeType: mimeType,
			Text:     text.String(),
		}
	}

	r.requests[e.RequestID] = entry
	entry.Pageref = pageID
}

func (r *Recorder) onResponse(e *network.EventResponseReceived) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.requests[e.RequestID]
	if !ok {
		return
	}

	headers := make([]*har.NameValuePair, 0, len(e.Response.Headers))
	for k, v := range e.Response.Headers {
		s, _ := v.(string)
		headers = append(headers, &har.NameValuePair{Name: k, Value: s})
	}

	entry.Response = &har.Response{
		Status:      int64(e.Response.Status),
		StatusText:  e.Response.StatusText,
		HTTPVersion: e.Response.Protocol,
		Headers:     headers,
		Content: &har.Content{
			Size:     int64(e.Response.EncodedDataLength),
			MimeType: e.Response.MimeType,
		},
	}
}

func (r *Recorder) onLoadingFinished(ctx context.Context, e *network.EventLoadingFinished) {
	r.mu.Lock()
	entry, ok := r.requests[e.RequestID]
	r.mu.Unlock()

	if !ok {
		return
	}

	body, err := network.GetResponseBody(e.RequestID).Do(ctx)
	if err == nil {
		entry.Response.Content.Text = base64.StdEncoding.EncodeToString(body)
		entry.Response.Content.Encoding = "base64"
	}

	r.mu.Lock()
	r.entries = append(r.entries, entry)
	delete(r.requests, e.RequestID)
	r.mu.Unlock()
}

func (r *Recorder) onLoadingFailed(e *network.EventLoadingFailed) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Still record entry so failed requests appear in HAR
	// TODO: perhaps this should be a flag option
	if entry, ok := r.requests[e.RequestID]; ok {
		r.entries = append(r.entries, entry)
		delete(r.requests, e.RequestID)
	}
}

// Returns collected HAR entries
func (r *Recorder) Entries() []har.Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]har.Entry, len(r.entries))
	for i, e := range r.entries {
		result[i] = *e
	}
	return result
}

func (r *Recorder) Pages() []har.Page {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]har.Page, 0, len(r.pages))
	for _, p := range r.pages {
		result = append(result, *p)
	}
	return result
}

func headerValue(h network.Headers, name string) (string, bool) {
	for k, v := range h {
		if strings.EqualFold(k, name) {
			s, ok := v.(string)
			return s, ok
		}
	}
	return "", false
}
