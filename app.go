package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	width  = 480
	height = 1080
)

type App struct {
	background *ebiten.Image
}

func NewApp() (*App, error) {
	background, _, err := ebitenutil.NewImageFromFile("assets/background.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	return &App{
		background: background,
	}, nil
}

func (app *App) Update(screen *ebiten.Image) error {
	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
}

func (app *App) Layout(w, h int) (int, int) {
	return width, height
}
