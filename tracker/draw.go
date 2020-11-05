package tracker

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
)

func (tracker *Tracker) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	drawState := func(state bool, sheet *ebiten.Image) {
		for k := range tracker.items {
			if tracker.items[k].Enabled != state {
				continue
			}

			pos := tracker.items[k].Rect().Min.Add(tracker.pos)
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

func (tracker *Tracker) drawDungeons(screen *ebiten.Image) {
	for k := range tracker.items {
		if !tracker.items[k].IsMedallion {
			continue
		}

		rect := tracker.items[k].Rect()
		x, y := rect.Min.X, rect.Max.Y
		text.Draw(screen, tracker.items[k].DungeonText(), tracker.fontSmall, x, y, color.White)
	}
}

func (tracker *Tracker) drawCapacities(screen *ebiten.Image) {
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
		text.Draw(screen, str, tracker.font, x, y, color.White)
	}
}

func (tracker *Tracker) drawInputState(screen *ebiten.Image) {
	pos := tracker.pos.Add(image.Point{10, 15 + 9*gridSize})
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
		if tracker.input.textInputFor == hintTypeWOTH ||
			tracker.input.textInputFor == hintTypeBarren {
			if match := tracker.matchLocation(string(tracker.input.buf)); match != "" {
				str += " (" + match + ")"
			}
		} else if tracker.input.textInputFor == hintTypeAlways {
			index, _ := tracker.parseAlways(string(tracker.input.buf))
			if index > -1 {
				str += fmt.Sprintf(` (%s)`, tracker.getAlwaysLocations()[index])
			}
		}

	case inputStateDungeonInput:
		str = "dungeon for: "
		idx := tracker.getItemIndexByName(
			tracker.dungeonInputMedallionOrder[tracker.input.curMedallion],
		)
		str += tracker.items[idx].Name

		// Highlight corresponding medallion.
		rect := tracker.items[idx].Rect()
		size := rect.Size()
		ebitenutil.DrawRect(
			screen,
			float64(rect.Min.X), float64(rect.Min.Y),
			float64(size.X), float64(size.Y),
			color.RGBA{0xFF, 0xFF, 0xFF, 0x50},
		)
	}

	if str == "" {
		return
	}

	text.Draw(screen, str, tracker.fontSmall, pos.X, pos.Y, color.White)
}

func (tracker *Tracker) drawHints(screen *ebiten.Image) {
	lineHeight := tracker.hintSize.Y / maxHintsPerRow
	pos := tracker.hintPos.Add(margins)
	op := ebiten.DrawImageOptions{}

	for k, v := range tracker.getDrawableHintList() {
		if k > 0 && k%maxHintsPerRow == 0 {
			pos = pos.Add(image.Point{tracker.hintSize.X / 2, -maxHintsPerRow * lineHeight})
		}

		ebitenutil.DrawRect(
			screen,
			float64(pos.X-margins.X), float64(pos.Y-margins.Y),
			float64(tracker.hintSize.X/2),
			float64(tracker.hintSize.Y/maxHintsPerRow),
			v.bgColor,
		)

		if v.gfx != nil {
			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(pos.X),
				float64(pos.Y-margins.Y),
			)

			screen.DrawImage(tracker.sheetEnabled.SubImage(*v.gfx).(*ebiten.Image), &op)
			text.Draw(screen, v.text, tracker.fontSmall, pos.X+25, pos.Y, color.Black)
		} else {
			text.Draw(screen, v.text, tracker.fontSmall, pos.X, pos.Y, color.Black)
		}

		pos.Y += lineHeight
	}
}

type drawableHintEntry struct {
	text    string
	gfx     *image.Rectangle
	bgColor color.RGBA
}

var margins = image.Point{3, 15}

const (
	maxHintsPerRow = 10
)

func (tracker *Tracker) getDrawableHintList() []drawableHintEntry {
	entries := make(
		[]drawableHintEntry, 0,
		len(tracker.woths)+len(tracker.barrens)+
			len(tracker.sometimes)+len(tracker.always),
	)

	for _, v := range tracker.woths {
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

		entries = append(entries, drawableHintEntry{
			text:    v,
			bgColor: color.RGBA{255, 230, 153, 0xFF},
			gfx: &image.Rectangle{
				tracker.alwaysHints[name],
				image.Point{
					tracker.alwaysHints[name].X + itemSpriteWidth,
					tracker.alwaysHints[name].Y + itemSpriteHeight,
				},
			},
		})
	}

	return entries
}
