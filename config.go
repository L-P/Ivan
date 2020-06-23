package main

import (
	"encoding/json"
	"image"
	"ivan/tracker"
	"os"
)

type config struct {
	Items       []tracker.Item
	ZoneItemMap [9][9]string
	Dimensions  struct {
		ItemTracker image.Rectangle
		Timer       image.Rectangle
	}
}

func loadConfig(path string) (config, error) {
	f, err := os.Open(path)
	if err != nil {
		return config{}, err
	}
	defer f.Close()

	var ret config
	dec := json.NewDecoder(f)
	if err := dec.Decode(&ret); err != nil {
		return config{}, err
	}

	return ret, nil
}
