package harutil

import (
	"github.com/chromedp/cdproto/har"
	"github.com/chromedp/cdproto/network"
)

func ConvertNetworkHeaders(h network.Headers) []*har.NameValuePair {
	out := make([]*har.NameValuePair, 0, len(h))
	for k, v := range h {
		out = append(out, &har.NameValuePair{Name: k, Value: v.(string)})
	}
	return out
}
