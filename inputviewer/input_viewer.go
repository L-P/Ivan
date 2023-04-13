package inputviewer

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	btnSpriteWidth  = 34
	btnSpriteHeight = btnSpriteWidth
)

type Config struct {
	Enabled                   bool
	A, B, Z, R, Start         InputButton
	CUp, CRight, CDown, CLeft InputButton

	J InputJoystick
}

type InputButton struct {
	Type string // (Button, Axis)
	ID   int
	Dir  int // -1 / 1

	Pos, SheetPos image.Point

	Color color.RGBA
}

// Rect returns the position of the button gfx relative to the background origin.
func (ib InputButton) Rect() image.Rectangle {
	dX, dY := btnSpriteWidth/2, btnSpriteHeight/2

	return image.Rect(
		ib.Pos.X-dX, ib.Pos.Y-dX,
		ib.Pos.X+btnSpriteWidth-dX, ib.Pos.Y+btnSpriteHeight-dY,
	)
}

// SheetRect returns the position of the button gfx sprite on the spritesheet.
func (ib InputButton) SheetRect() image.Rectangle {
	return image.Rect(
		ib.SheetPos.X, ib.SheetPos.Y,
		ib.SheetPos.X+btnSpriteWidth, ib.SheetPos.Y+btnSpriteHeight,
	)
}

type InputJoystick struct {
	IDX, IDY int

	Pos, SheetPos image.Point
	Color         color.RGBA
}

// Rect returns the position of the joystick gfx relative to the background origin.
func (ij InputJoystick) Rect() image.Rectangle {
	dX, dY := btnSpriteWidth/2, btnSpriteHeight/2

	return image.Rect(
		ij.Pos.X-dX, ij.Pos.Y-dX,
		ij.Pos.X+btnSpriteWidth-dX, ij.Pos.Y+btnSpriteHeight-dY,
	)
}

// SheetRect returns the position of the joystick gfx sprite on the spritesheet.
func (ij InputJoystick) SheetRect() image.Rectangle {
	return image.Rect(
		ij.SheetPos.X, ij.SheetPos.Y,
		ij.SheetPos.X+btnSpriteWidth, ij.SheetPos.Y+btnSpriteHeight,
	)
}

func (ij InputJoystick) axes(id ebiten.GamepadID) (float64, float64) {
	return ebiten.GamepadAxis(id, ij.IDX), ebiten.GamepadAxis(id, ij.IDY)
}

func (ib InputButton) pressed(id ebiten.GamepadID) bool {
	switch ib.Type {
	case "Button":
		return ebiten.IsGamepadButtonPressed(id, ebiten.GamepadButton(ib.ID))
	case "Axis":
		if ib.Dir < 0 {
			return ebiten.GamepadAxis(id, ib.ID) < -0.15
		}
		return ebiten.GamepadAxis(id, ib.ID) > 0.15
	}

	return false
}

type InputViewer struct {
	config Config
	id     ebiten.GamepadID

	sheetEnabled, sheetDisabled *ebiten.Image
}

func NewInputViewer(config Config) *InputViewer {
	if !config.Enabled {
		return nil
	}

	ids := ebiten.GamepadIDs()
	if len(ids) == 0 {
		return nil
	}
	id := ids[0]

	log.Printf("info: reading input from %s", ebiten.GamepadName(id))

	iv := &InputViewer{
		config: config,
		id:     id,
	}

	images := []struct {
		img  **ebiten.Image
		path string
	}{
		// HACK resources are loaded twice
		{&iv.sheetDisabled, "assets/items-disabled.png"},
		{&iv.sheetEnabled, "assets/items.png"},
	}

	var err error
	for _, v := range images {
		*v.img, _, err = ebitenutil.NewImageFromFile(v.path)
		if err != nil {
			log.Fatal(err)
		}
	}

	return iv
}

func (iv *InputViewer) Draw(screen *ebiten.Image) {
	if iv == nil { // allow ignoring input viewer
		return
	}

	iv.legacyDraw(screen)
	iv.drawButtons(screen)
	iv.drawJoystick(screen)
}

func (iv *InputViewer) drawJoystick(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}

	j := iv.config.J
	op.GeoM.Reset()
	op.GeoM.Translate(float64(j.Rect().Min.X), float64(j.Rect().Min.Y))
	screen.DrawImage(
		iv.sheetEnabled.SubImage(j.SheetRect()).(*ebiten.Image),
		&op,
	)
	x, y := j.axes(iv.id)
	ebitenutil.DrawRect(
		screen,
		float64(j.Pos.X)+14.0*x-1,
		float64(j.Pos.Y)+14.0*y-1,
		2, 2,
		color.RGBA{j.Color.R, j.Color.G, j.Color.B, 0xFF},
	)
}

func (iv *InputViewer) drawButtons(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}

	for _, v := range []*InputButton{
		&iv.config.A, &iv.config.B,
		&iv.config.Start,
		&iv.config.CUp, &iv.config.CRight,
		&iv.config.CDown, &iv.config.CLeft,
	} {
		sheet := iv.sheetEnabled
		if !v.pressed(iv.id) {
			sheet = iv.sheetDisabled
		}

		op.GeoM.Reset()
		op.GeoM.Translate(float64(v.Rect().Min.X), float64(v.Rect().Min.Y))
		screen.DrawImage(
			sheet.SubImage(v.SheetRect()).(*ebiten.Image),
			&op,
		)
	}
}

// TODO replace when a better idea is found for Z/R.
func (iv *InputViewer) legacyDraw(screen *ebiten.Image) {
	for _, v := range []*InputButton{
		&iv.config.Z, &iv.config.R,
	} {
		if !v.pressed(iv.id) {
			continue
		}

		ebitenutil.DrawRect(
			screen,
			float64(v.Pos.X), float64(v.Pos.Y),
			2, 2,
			color.RGBA{v.Color.R, v.Color.G, v.Color.B, 0xFF},
		)
	}
}
