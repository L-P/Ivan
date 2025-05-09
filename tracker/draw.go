package tracker

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (tracker *Tracker) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	drawState := func(state bool, sheet *ebiten.Image) {
		for k := range tracker.items {
			if tracker.items[k].Enabled != state {
				continue
			}

			pos := tracker.cfg.Layout.ItemTracker.Min.Add(
				tracker.items[k].Rect().Min,
			)
			op.GeoM.Reset()
			op.GeoM.Translate(float64(pos.X), float64(pos.Y))

			screen.DrawImage(
				sheet.SubImage(tracker.items[k].SheetRect()).(*ebiten.Image),
				&op,
			)
		}
	}

	screen.DrawImage(tracker.background, nil)
	if tracker.kbInputStateIsAny(inputStateItemKPZoneInput, inputStateItemInput) {
		screen.DrawImage(tracker.backgroundHelp, nil)
		if tracker.input.activeKPZone > 0 {
			tracker.drawActiveItemSlot(screen, tracker.input.activeKPZone)
		}
	}

	// Do two loops to avoid texture switches.
	drawState(false, tracker.sheetDisabled)
	drawState(true, tracker.sheetEnabled)

	tracker.drawDungeons(screen)
	tracker.drawCapacities(screen)
	tracker.drawInputState(screen)
	tracker.drawHints(screen)
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

	switch slot {
	case 2: // KP 9
		pos.Y = 0
		size.X = gridSize
		size.Y = 4*gridSize + (gridSize / 2)
	case 8: // KP 3
		pos.Y = 4*gridSize + (gridSize / 2)
		size.X = gridSize
		size.Y = pos.Y
	}

	vector.DrawFilledRect(
		screen,
		float32(pos.X), float32(pos.Y),
		float32(size.X), float32(size.Y),
		color.RGBA{0xFF, 0xFF, 0xFF, 0x50},
		false,
	)
}

func (tracker *Tracker) drawDungeons(screen *ebiten.Image) {
	var op = &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(color.White)

	for k := range tracker.items {
		if !tracker.items[k].IsMedallion {
			continue
		}

		rect := tracker.items[k].Rect()

		op.GeoM.Reset()
		op.GeoM.Translate(float64(rect.Min.X), float64(rect.Max.Y-trackerSmallFontSize))
		text.Draw(screen, tracker.items[k].DungeonText(), tracker.fontSmall, op)
	}
}

func (tracker *Tracker) drawCapacities(screen *ebiten.Image) {
	var op = &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(color.White)

	for k := range tracker.items {
		var count int
		switch {
		case tracker.items[k].HasCapacity():
			count = tracker.items[k].Capacity()
		case tracker.items[k].IsCountable():
			count = tracker.items[k].Count
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
		op.GeoM.Reset()
		op.GeoM.Translate(float64(x), float64(y-trackerFontSize))
		text.Draw(screen, str, tracker.font, op)
	}
}

func (tracker *Tracker) drawInputState(screen *ebiten.Image) {
	var str string

	switch tracker.input.state {
	case inputStateIdle:
		// NOP
	case inputStateItemInput, inputStateItemKPZoneInput:
		if tracker.input.downgradeNextItem {
			str = "-"
		} else {
			str = "+"
		}

	case inputStateTextInput:
		str = "> " + string(tracker.input.buf)
		switch tracker.input.textInputFor { //nolint:exhaustive
		case hintTypeWOTH, hintTypeBarren:
			if match := tracker.matchLocation(string(tracker.input.buf)); match != "" {
				str += " (" + match + ")"
			}
		case hintTypeAlways:
			index, _ := tracker.parseAlways(string(tracker.input.buf))
			if index > -1 {
				str += fmt.Sprintf(` (%s)`, tracker.getAlwaysLocations()[index])
			}
		}

	case inputStateDungeonInput:
		str = "dungeon for: "
		idx := tracker.getItemIndexByName(
			tracker.cfg.ItemTracker.DungeonInputMedallionOrder[tracker.input.curMedallion],
		)
		str += tracker.items[idx].Name

		// Highlight corresponding medallion.
		rect := tracker.items[idx].Rect()
		size := rect.Size()
		vector.DrawFilledRect(
			screen,
			float32(rect.Min.X), float32(rect.Min.Y),
			float32(size.X), float32(size.Y),
			color.RGBA{0xFF, 0xFF, 0xFF, 0x50},
			false,
		)
	}

	if str == "" {
		return
	}

	pos := tracker.cfg.Layout.ItemTracker.Min.Add(
		image.Point{10, 15 + 9*gridSize},
	)

	op := &text.DrawOptions{}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Reset()
	op.GeoM.Translate(float64(pos.X), float64(pos.Y)-trackerSmallFontSize)
	text.Draw(screen, str, tracker.fontSmall, op)
}

func (tracker *Tracker) drawHints(screen *ebiten.Image) {
	var (
		margins     = image.Point{3, 15}
		iconOffsetY = 1
		size        = tracker.cfg.Layout.HintTracker.Size()
		pos         = tracker.cfg.Layout.HintTracker.Min.Add(margins)
		lineHeight  = size.Y / maxHintsPerRow
		op          = ebiten.DrawImageOptions{}
		textOp      = &text.DrawOptions{}
	)
	textOp.ColorScale.ScaleWithColor(color.Black)

	for k, v := range tracker.getDrawableHintList() {
		if k > 0 && k%maxHintsPerRow == 0 {
			pos = pos.Add(image.Point{size.X / 2, -maxHintsPerRow * lineHeight})
		}

		vector.DrawFilledRect(
			screen,
			float32(pos.X-margins.X), float32(pos.Y-margins.Y),
			float32(size.X/2),
			float32(size.Y/maxHintsPerRow),
			v.bgColor,
			false,
		)

		textOp.GeoM.Reset()
		textOp.GeoM.Translate(float64(pos.X), float64(pos.Y)-trackerSmallFontSize)

		if v.gfx != nil {
			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(pos.X),
				float64(pos.Y-margins.Y+iconOffsetY),
			)

			screen.DrawImage(tracker.sheetEnabled.SubImage(*v.gfx).(*ebiten.Image), &op)
			textOp.GeoM.Translate(25, 0)
		}

		text.Draw(screen, v.text, tracker.fontSmall, textOp)

		pos.Y += lineHeight
	}
}

type drawableHintEntry struct {
	text    string
	gfx     *image.Rectangle
	bgColor color.RGBA
}

const (
	maxHintsPerRow = 10
)

func (tracker *Tracker) getDrawableHintList() []drawableHintEntry {
	entries := make(
		[]drawableHintEntry, 0,
		len(tracker.woths)+
			len(tracker.barrens)+
			len(tracker.sometimes)+
			len(tracker.always)+
			len(tracker.goals),
	)

	for _, v := range tracker.woths {
		entries = append(entries, drawableHintEntry{text: v, bgColor: color.RGBA{212, 234, 107, 0xFF}})
	}

	for _, v := range tracker.goals {
		entries = append(entries, drawableHintEntry{text: v, bgColor: color.RGBA{212, 234, 107, 0xFF}})
	}

	for _, v := range tracker.barrens {
		entries = append(entries, drawableHintEntry{text: v, bgColor: color.RGBA{255, 109, 109, 0xFF}})
	}

	for _, v := range tracker.sometimes {
		entries = append(entries, drawableHintEntry{text: v, bgColor: color.RGBA{180, 198, 231, 0xFF}})
	}

	for k, v := range tracker.always {
		name := tracker.getAlwaysLocations()[k]
		if v == "" {
			continue
		}

		entry := drawableHintEntry{
			text:    v,
			bgColor: color.RGBA{255, 230, 153, 0xFF},
			gfx: &image.Rectangle{
				tracker.cfg.HintTracker.AlwaysHints[name],
				image.Point{
					tracker.cfg.HintTracker.AlwaysHints[name].X + itemSpriteWidth,
					tracker.cfg.HintTracker.AlwaysHints[name].Y + itemSpriteHeight,
				},
			},
		}

		entries = append(entries, entry)
	}

	return entries
}
