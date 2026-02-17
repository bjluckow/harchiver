package harutil

import (
	"encoding/json"
	"io"
	"os"

	"github.com/chromedp/cdproto/har"
)

func Parse(path string) (*har.HAR, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var archive har.HAR
	if err := json.Unmarshal(data, &archive); err != nil {
		return nil, err
	}
	return &archive, nil
}

func Write(archive *har.HAR, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(archive)
}
