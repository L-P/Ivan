package tracker

import (
	"image"
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
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

			if err := screen.DrawImage(
				sheet.SubImage(tracker.items[k].SheetRect()).(*ebiten.Image),
				&op,
			); err != nil {
				log.Fatal(err)
			}
		}
	}

	_ = screen.DrawImage(tracker.background, nil)
	if tracker.kbInputStateIsAny(inputStateItemKPZoneInput, inputStateItemInput) {
		_ = screen.DrawImage(tracker.backgroundHelp, nil)
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
