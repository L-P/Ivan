package tracker

import (
	"encoding/json"
	"os"
)

type Config struct {
	Items []Item

	ZoneItems [9][9]string
}

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var ret Config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&ret); err != nil {
		return Config{}, err
	}

	return ret, nil
}
