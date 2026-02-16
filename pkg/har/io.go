package har

import (
	"encoding/json"
	"io"
	"os"
)

func Parse(path string) (*HttpArchive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var archive HttpArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		return nil, err
	}
	return &archive, nil
}

func Write(archive *HttpArchive, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(archive)
}
