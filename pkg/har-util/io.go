package harutil

import (
	"encoding/json"
	"io"
	"os"

	hartype "github.com/bjluckow/harchiver/pkg/har-type"
)

func Parse(path string) (*hartype.HttpArchive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var archive hartype.HttpArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		return nil, err
	}
	return &archive, nil
}

func Write(archive *hartype.HttpArchive, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(archive)
}
