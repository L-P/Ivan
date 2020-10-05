package inputviewer

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type Config struct {
	A, B, Z, R, Start         InputButton
	CUp, CRight, CDown, CLeft InputButton

	J InputJoystick
}

type InputButton struct {
	Type string // (Button, Axis)
	ID   int
	Dir  int // -1 / 1

	Pos   image.Point
	Color color.RGBA
}

type InputJoystick struct {
	IDX, IDY int

	Pos   image.Point
	Color color.RGBA
}

func (i InputJoystick) axes(id int) (float64, float64) {
	return ebiten.GamepadAxis(id, i.IDX), ebiten.GamepadAxis(id, i.IDY)
}

func (i InputButton) pressed(id int) bool {
	dir := 1.0
	if dir < 0 {
		dir = -1
	}

	switch i.Type {
	case "Button":
		return ebiten.IsGamepadButtonPressed(id, ebiten.GamepadButton(i.ID))
	case "Axis":
		return ebiten.GamepadAxis(id, i.ID) > (dir * 0.15)
	}

	return false
}

type InputViewer struct {
	config Config
	id     int
}

func NewInputViewer(config Config) *InputViewer {
	ids := ebiten.GamepadIDs()
	if len(ids) == 0 {
		log.Printf("warning: no gamepad found")
		return nil
	}
	id := ids[0]

	log.Printf("info: reading input from %s", ebiten.GamepadName(id))

	return &InputViewer{
		config: config,
		id:     id,
	}
}

func (iv *InputViewer) Draw(screen *ebiten.Image) {
	if iv == nil { // allow ignoring input viewer
		return
	}

	for _, v := range []*InputButton{
		&iv.config.A, &iv.config.B,
		&iv.config.Z, &iv.config.R,
		&iv.config.Start,
		&iv.config.CUp, &iv.config.CRight,
		&iv.config.CDown, &iv.config.CLeft,
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

	j := iv.config.J
	x, y := j.axes(iv.id)
	ebitenutil.DrawRect(
		screen,
		float64(j.Pos.X)+4.0*x,
		float64(j.Pos.Y)+4.0*y,
		2, 2,
		color.RGBA{j.Color.R, j.Color.G, j.Color.B, 0xFF},
	)
}
