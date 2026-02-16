package harutil

import (
	hartype "github.com/bjluckow/harchiver/pkg/har-type"
	"github.com/chromedp/cdproto/network"
)

func ConvertNetworkHeaders(h network.Headers) []hartype.Header {
	out := make([]hartype.Header, 0, len(h))
	for k, v := range h {
		out = append(out, hartype.Header{Name: k, Value: v.(string)})
	}
	return out
}
