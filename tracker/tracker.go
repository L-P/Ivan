package tracker

import (
	"encoding/json"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

type Tracker struct {
	pos, size         image.Point
	hintPos, hintSize image.Point

	background, backgroundHelp  *ebiten.Image
	sheetDisabled, sheetEnabled *ebiten.Image
	font, fontSmall             font.Face

	items       []Item
	zoneItemMap ZoneItemMap
	locations   []string
	input       kbInput
	binds       map[string]string
	alwaysHints map[string]image.Point // name => sheet pos

	woths, barrens, sometimes []string
	always                    [7]string // in order: skull, bigg, 30, 40, 50, OOT, frogs 2

	dungeonInputMedallionOrder, dungeonInputDungeonKP []string

	undoStack, redoStack []undoStackEntry
}

const (
	capacityFontSize = 20
	templeFontSize   = 13
)

type ZoneItemMap [9][9]string

func New( // TODO pass config
	dimensions image.Rectangle,
	hintDimensions image.Rectangle,
	items []Item,
	zoneItemMap ZoneItemMap,
	locations []string,
	binds map[string]string,
	alwaysHints map[string]image.Point,
	dungeonInputMedallionOrder []string,
	dungeonInputDungeonKP []string,
) (*Tracker, error) {
	tracker := &Tracker{
		pos:      dimensions.Min,
		size:     dimensions.Size(),
		hintPos:  hintDimensions.Min,
		hintSize: hintDimensions.Size(),

		dungeonInputMedallionOrder: dungeonInputMedallionOrder,
		dungeonInputDungeonKP:      dungeonInputDungeonKP,

		locations:   locations,
		binds:       binds,
		alwaysHints: alwaysHints,
		zoneItemMap: zoneItemMap,
	}

	tracker.items = make([]Item, len(items))
	copy(tracker.items, items)

	if err := tracker.loadResources(); err != nil {
		return nil, err
	}

	tracker.setInitialItems()

	return tracker, nil
}

func (tracker *Tracker) loadResources() (err error) {
	images := []struct {
		img  **ebiten.Image
		path string
	}{
		{&tracker.background, "assets/background.png"},
		{&tracker.backgroundHelp, "assets/background-help.png"},
		{&tracker.sheetDisabled, "assets/items-disabled.png"},
		{&tracker.sheetEnabled, "assets/items.png"},
	}

	for _, v := range images {
		*v.img, _, err = ebitenutil.NewImageFromFile(v.path)
		if err != nil {
			return err
		}
	}

	ttf, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}

	tracker.font = truetype.NewFace(ttf, &truetype.Options{
		Size:    capacityFontSize,
		Hinting: font.HintingFull,
	})
	tracker.fontSmall = truetype.NewFace(ttf, &truetype.Options{
		Size:    templeFontSize,
		Hinting: font.HintingFull,
	})

	return nil
}

func (tracker *Tracker) GetZoneItem(zoneKP, itemKP int) (string, error) {
	if zoneKP <= 0 || zoneKP > 9 {
		return "", errInvalidZone{zoneKP, itemKP}
	}
	if itemKP <= 0 || itemKP > 9 {
		return "", errInvalidZone{zoneKP, itemKP}
	}

	name := tracker.zoneItemMap[zoneKP-1][itemKP-1]
	if name == "" {
		return "", errNoDefinition{zoneKP, itemKP}
	}

	return name, nil
}

func (tracker *Tracker) GetZoneDungeon(zoneKP int) (string, error) {
	if zoneKP <= 0 || zoneKP > 9 {
		return "", errInvalidZone{zoneKP, 0}
	}

	dungeon := tracker.dungeonInputDungeonKP[zoneKP-1]
	if dungeon == "" {
		return "", errNoDefinition{-1, zoneKP}
	}

	return dungeon, nil
}

func (tracker *Tracker) GetZoneItemIndex(zoneKP, itemKP int) (int, error) {
	name, err := tracker.GetZoneItem(zoneKP, itemKP)
	if err != nil {
		return -1, err
	}

	index := tracker.getItemIndexByName(name)
	if index < 0 {
		return -1, errInvalidDefinition{zoneKP, itemKP, name}
	}

	return index, nil
}

// getItemIndexByPos returns the index of the item at the given position or -1 if
// there is no item under the given pixel.
func (tracker *Tracker) getItemIndexByPos(x, y int) int {
	for k := range tracker.items {
		if (image.Point{x, y}).In(tracker.items[k].Rect()) {
			return k
		}
	}

	return -1
}

func (tracker *Tracker) getItemIndexByName(name string) int {
	for k := range tracker.items {
		if tracker.items[k].Name == name {
			return k
		}
	}

	return -1
}

// ClickLeft upgrades the item under the given point.
func (tracker *Tracker) ClickLeft(x, y int) {
	i := tracker.getItemIndexByPos(x, y)
	if i < 0 {
		return
	}

	tracker.changeItem(i, true)
}

// ClickRight downgrades the item under the given point.
func (tracker *Tracker) ClickRight(x, y int) {
	i := tracker.getItemIndexByPos(x, y)
	if i < 0 {
		return
	}

	tracker.changeItem(i, false)
}

func (tracker *Tracker) changeItem(itemIndex int, isUpgrade bool) {
	var fn func() bool
	if isUpgrade {
		fn = tracker.items[itemIndex].Upgrade
	} else {
		fn = tracker.items[itemIndex].Downgrade
	}

	if fn() {
		tracker.appendToUndoStack(itemIndex, isUpgrade)
	}
}

func (tracker *Tracker) Wheel(x, y int, up bool) {
	i := tracker.getItemIndexByPos(x, y)
	if i < 0 {
		return
	}

	switch {
	case tracker.items[i].IsMedallion:
		tracker.items[i].CycleDungeon(up)
	default:
		if up {
			tracker.ClickLeft(x, y)
		} else {
			tracker.ClickRight(x, y)
		}
	}
}

func (tracker *Tracker) Reset(items []Item, zoneItemMap ZoneItemMap) {
	tracker.items = make([]Item, len(items))
	copy(tracker.items, items)

	tracker.zoneItemMap = zoneItemMap
	tracker.undoStack = tracker.undoStack[:0]
	tracker.redoStack = tracker.redoStack[:0]
	tracker.woths = tracker.woths[:0]
	tracker.barrens = tracker.barrens[:0]
	tracker.sometimes = tracker.sometimes[:0]
	tracker.always = [7]string{}
	tracker.setInitialItems()

	if err := tracker.Save(); err != nil {
		log.Printf("error: %s", err)
	}
}

func (tracker *Tracker) Save() error {
	f, err := os.OpenFile(getSavePath(), os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(tracker)
}

func (tracker *Tracker) Load() error {
	f, err := os.Open(getSavePath())
	if err != nil {
		return err
	}
	defer f.Close()

	return tracker.loadJSON(f)
}

func getSavePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "./"
	}

	return filepath.Join(dir, "ivan.state.json")
}

func (tracker Tracker) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Items                     []Item
		WotHs, Barrens, Sometimes []string
		Always                    [7]string
		UndoStack, RedoStack      []undoStackEntry
	}{
		tracker.items,
		tracker.woths,
		tracker.barrens,
		tracker.sometimes,
		tracker.always,
		tracker.undoStack,
		tracker.redoStack,
	})
}

func (tracker *Tracker) loadJSON(r io.Reader) error {
	var tmp struct {
		Items                     []Item
		WotHs, Barrens, Sometimes []string
		Always                    [7]string
		UndoStack, RedoStack      []undoStackEntry
	}

	dec := json.NewDecoder(r)
	if err := dec.Decode(&tmp); err != nil {
		return err
	}

	tracker.items = tmp.Items
	tracker.woths = tmp.WotHs
	tracker.barrens = tmp.Barrens
	tracker.sometimes = tmp.Sometimes
	tracker.always = tmp.Always
	tracker.undoStack = tmp.UndoStack
	tracker.redoStack = tmp.RedoStack

	return nil
}

func (tracker *Tracker) setInitialItems() {
	tracker.changeItem(tracker.getItemIndexByName("Gold Skulltula Token"), true)
	tracker.changeItem(tracker.getItemIndexByName("Kokiri Tunic"), true)
	tracker.changeItem(tracker.getItemIndexByName("Kokiri Boots"), true)
	tracker.changeItem(tracker.getItemIndexByName("Master Sword"), true)

	// OOTR S4 settings.
	tracker.changeItem(tracker.getItemIndexByName("Deku Shield"), true)
	tracker.changeItem(tracker.getItemIndexByName("Deku Nut"), true)
	tracker.changeItem(tracker.getItemIndexByName("Deku Stick"), true)
	tracker.changeItem(tracker.getItemIndexByName("Ocarina"), true)
	for i := 0; i < 3; i++ { // get Zelda's Letter
		tracker.changeItem(tracker.getItemIndexByName("Mask Trade Sequence"), true)
	}
}
