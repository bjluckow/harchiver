package har

import (
	"encoding/json"
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

func Write(archive *HttpArchive, path string) error {
	data, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
