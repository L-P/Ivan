package main

import (
	"errors"
	"ivan/tracker"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	width  = 296
	height = 393
)

var errCloseApp = errors.New("user requested app close")

type App struct {
	background *ebiten.Image
	tracker    *tracker.Tracker

	// For debouncing mouse clicks.
	lastMouseState bool
}

func NewApp() (*App, error) {
	background, _, err := ebitenutil.NewImageFromFile("assets/background.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	tracker, err := tracker.New("assets/items.json")
	if err != nil {
		return nil, err
	}

	return &App{
		background: background,
		tracker:    tracker,
	}, nil
}

func (app *App) Update(screen *ebiten.Image) error {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyEscape):
		return errCloseApp
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
		if !app.lastMouseState {
			app.tracker.Upgrade(ebiten.CursorPosition())
		}
		app.lastMouseState = true
	case ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight):
		if !app.lastMouseState {
			app.tracker.Downgrade(ebiten.CursorPosition())
		}
		app.lastMouseState = true
	default:
		app.lastMouseState = false
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	if err := screen.DrawImage(app.background, nil); err != nil {
		log.Fatal(err)
	}

	app.tracker.Draw(screen)
}

func (app *App) Layout(w, h int) (int, int) {
	return width, height
}
