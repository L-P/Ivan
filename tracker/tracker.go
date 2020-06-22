package tracker

import (
	"encoding/json"
	"image"
	"image/color"
	"log"
	"os"
	"strconv"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

type Tracker struct {
	Origin image.Point

	background     *ebiten.Image
	backgroundHelp *ebiten.Image
	font           font.Face
	fontSmall      font.Face
	sheetDisabled  *ebiten.Image
	sheetEnabled   *ebiten.Image

	items []Item
	input kbInput

	undoStack  []undoStackEntry
	undoCursor int
}

// undoStackEntry represents an action (upgrade/downgrade) that happened on an item.
type undoStackEntry struct {
	itemIndex int
	isUpgrade bool
}

const (
	Width  = 7 * gridSize
	Height = 9 * gridSize

	capacityFontSize = 20
	templeFontSize   = 13
)

func New(path string) (*Tracker, error) {
	background, _, err := ebitenutil.NewImageFromFile("assets/background.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	backgroundHelp, _, err := ebitenutil.NewImageFromFile("assets/background-help.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	items, err := loadItems(path)
	if err != nil {
		return nil, err
	}

	ttf, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	sheetDisabled, _, err := ebitenutil.NewImageFromFile("assets/items-disabled.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	sheetEnabled, _, err := ebitenutil.NewImageFromFile("assets/items.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	tracker := &Tracker{
		background:     background,
		backgroundHelp: backgroundHelp,
		items:          items,
		sheetDisabled:  sheetDisabled,
		sheetEnabled:   sheetEnabled,
		font: truetype.NewFace(ttf, &truetype.Options{
			Size:    capacityFontSize,
			Hinting: font.HintingFull,
		}),
		fontSmall: truetype.NewFace(ttf, &truetype.Options{
			Size:    templeFontSize,
			Hinting: font.HintingFull,
		}),
	}

	return tracker, nil
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

func (tracker *Tracker) appendToUndoStack(itemIndex int, isUpgrade bool) {
	// If we were back in time, discard and replace history.
	if tracker.undoCursor > 0 && tracker.undoCursor != len(tracker.undoStack)-1 {
		tracker.undoStack = tracker.undoStack[:tracker.undoCursor+1]
	}

	tracker.undoStack = append(tracker.undoStack, undoStackEntry{
		itemIndex: itemIndex,
		isUpgrade: isUpgrade,
	})
	tracker.undoCursor = len(tracker.undoStack) - 1
}

func (tracker *Tracker) undo() {
	if tracker.undoCursor < 0 || len(tracker.undoStack) == 0 {
		log.Printf("no action to undo")
		return
	}

	entry := tracker.undoStack[tracker.undoCursor]
	if entry.isUpgrade {
		tracker.items[entry.itemIndex].Downgrade()
	} else {
		tracker.items[entry.itemIndex].Upgrade()
	}

	tracker.undoCursor--
}

func (tracker *Tracker) redo() {
	if tracker.undoCursor >= len(tracker.undoStack)-1 {
		log.Printf("no action to redo")
		return
	}

	entry := tracker.undoStack[tracker.undoCursor+1]
	if entry.isUpgrade {
		tracker.items[entry.itemIndex].Upgrade()
	} else {
		tracker.items[entry.itemIndex].Downgrade()
	}

	tracker.undoCursor++
}

func (tracker *Tracker) Wheel(x, y int, up bool) {
	i := tracker.getItemIndexByPos(x, y)
	if i < 0 {
		return
	}

	switch {
	case tracker.items[i].IsMedallion:
		tracker.items[i].CycleTemple(up)
	default:
		if up {
			tracker.ClickLeft(x, y)
		} else {
			tracker.ClickRight(x, y)
		}
	}
}

func (tracker *Tracker) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	drawState := func(state bool, sheet *ebiten.Image) {
		for k := range tracker.items {
			if tracker.items[k].Enabled != state {
				continue
			}

			pos := tracker.items[k].Rect().Min.Add(tracker.Origin)
			op.GeoM.Reset()
			op.GeoM.Translate(float64(pos.X), float64(pos.Y))

			if err := screen.DrawImage(
				sheet.SubImage(tracker.items[k].SheetRect()).(*ebiten.Image),
				&op,
			); err != nil {
				log.Fatal(err)
			}
		}
	}

	if tracker.kbInputStateIsAny(inputStateItemKPZoneInput, inputStateItemInput) {
		_ = screen.DrawImage(tracker.backgroundHelp, nil)
		if tracker.input.activeKPZone > 0 {
			tracker.drawActiveItemSlot(screen, tracker.input.activeKPZone)
		}
	} else {
		_ = screen.DrawImage(tracker.background, nil)
	}

	// Do two loops to avoid texture switches.
	drawState(false, tracker.sheetDisabled)
	drawState(true, tracker.sheetEnabled)

	tracker.drawTemples(screen)
	tracker.drawCapacities(screen)
	tracker.drawInputState(screen)
}

func (tracker *Tracker) drawActiveItemSlot(screen *ebiten.Image, slot int) {
	if slot <= 0 || slot > 9 {
		return
	}

	slot = []int{ // make maths ez
		0,
		6, 7, 8,
		3, 4, 5,
		0, 1, 2,
	}[slot]

	edge := gridSize * 3
	pos := image.Point{
		(slot % 3) * edge,
		(slot / 3) * edge,
	}
	size := image.Point{126, 126}

	if slot == 2 { // KP 9
		pos.Y = 0
		size.X = gridSize
		size.Y = 4*gridSize + (gridSize / 2)
	} else if slot == 8 { // KP 3
		pos.Y = 4*gridSize + (gridSize / 2)
		size.X = gridSize
		size.Y = pos.Y
	}

	ebitenutil.DrawRect(
		screen,
		float64(pos.X), float64(pos.Y),
		float64(size.X), float64(size.Y),
		color.RGBA{0xFF, 0xFF, 0xFF, 0x50},
	)
}

func (tracker *Tracker) drawTemples(screen *ebiten.Image) {
	for k := range tracker.items {
		if !tracker.items[k].IsMedallion {
			continue
		}

		rect := tracker.items[k].Rect()
		x, y := rect.Min.X, rect.Max.Y
		text.Draw(screen, tracker.items[k].TempleText(), tracker.fontSmall, x, y, color.White)
	}
}

func (tracker *Tracker) drawCapacities(screen *ebiten.Image) {
	for k := range tracker.items {
		var count int
		switch {
		case tracker.items[k].HasCapacity():
			count = tracker.items[k].Capacity()
		case tracker.items[k].IsCountable():
			count = tracker.items[k].Count()
		default:
			continue
		}

		if !tracker.items[k].Enabled {
			continue
		}

		rect := tracker.items[k].Rect()
		x, y := rect.Min.X, rect.Max.Y

		// HACK, display skull count centered on the right slot
		if tracker.items[k].Name == "Gold Skulltula Token" {
			x, y = rect.Min.X+gridSize+marginLeft, rect.Min.Y+marginTop+(gridSize/2)
			if count == 100 {
				x -= 5
			} else if count < 10 {
				x += 5
			}
		}

		str := strconv.Itoa(count)
		text.Draw(screen, str, tracker.font, x, y, color.White)
	}
}

func loadItems(path string) ([]Item, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []Item
	dec := json.NewDecoder(f)
	if err := dec.Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}
