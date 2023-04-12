package config

import (
	"encoding/json"
	"fmt"
	"image"
	inputviewer "ivan/input-viewer"
	"ivan/tracker"
	"os"
	"path/filepath"
)

type HintTracker struct {
	AlwaysHints map[string]image.Point
}

type ItemTracker struct {
	DungeonInputMedallionOrder []string
	DungeonInputDungeonKP      []string
	ZoneItemMap                [9][9]string
}

type Config struct {
	Binds       map[string]string
	HintTracker HintTracker
	ItemTracker ItemTracker

	Items     []tracker.Item
	Locations []string // regions and dungeons.

	InputViewer inputviewer.Config
	Dimensions  struct {
		ItemTracker image.Rectangle
		Timer       image.Rectangle
		HintTracker image.Rectangle
	}
}

func (c Config) WindowSize() image.Point {
	var ret image.Rectangle
	for _, v := range []image.Rectangle{
		c.Dimensions.ItemTracker,
		c.Dimensions.Timer,
		c.Dimensions.HintTracker,
	} {
		ret = ret.Union(v)
	}

	return ret.Size()
}

func NewFromDir(dir string) (Config, error) {
	var cfg Config
	src := map[string]interface{}{
		"binds.json":        &cfg.Binds,
		"hint_tracker.json": &cfg.HintTracker,
		"input_viewer.json": &cfg.InputViewer,
		"item_tracker.json": &cfg.ItemTracker,
		"items.json":        &cfg.Items,
		"layout.json":       &cfg.Dimensions,
		"locations.json":    &cfg.Locations,
	}

	for name, dst := range src {
		if err := unmarshalFile(dst, filepath.Join(dir, name)); err != nil {
			return Config{}, fmt.Errorf("unable to load '%s': %w", name, err)
		}
	}

	return cfg, nil
}

func unmarshalFile(dst interface{}, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open file '%s' for reading: %w", path, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("unable to decode json in '%s': %w", path, err)
	}

	return nil
}
