package main

import (
	"errors"
	"image"
	"ivan/timer"
	"ivan/tracker"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	width  = tracker.Width
	height = tracker.Height + timer.Height
)

var errCloseApp = errors.New("user requested app close")

type App struct {
	tracker *tracker.Tracker
	timer   *timer.Timer
}

func NewApp() (*App, error) {
	timer, err := timer.New(image.Point{0, tracker.Height})
	if err != nil {
		return nil, err
	}

	tracker, err := tracker.New("assets/config.json")
	if err != nil {
		return nil, err
	}

	return &App{
		tracker: tracker,
		timer:   timer,
	}, nil
}

func (app *App) Update(screen *ebiten.Image) error {
	_, wheel := ebiten.Wheel()

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		if !app.timer.IsRunning() {
			return errCloseApp
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyHome):
		if !app.timer.IsRunning() {
			app.tracker.Reset()
		}
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		app.timer.Toggle()
	case inpututil.IsKeyJustPressed(ebiten.KeyDelete):
		app.timer.Reset()
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		app.tracker.ClickLeft(ebiten.CursorPosition())
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		app.tracker.ClickRight(ebiten.CursorPosition())
	case wheel != 0:
		x, y := ebiten.CursorPosition()
		app.tracker.Wheel(x, y, wheel > 0)
	default:
		app.tracker.Input(ebiten.InputChars())
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	app.tracker.Draw(screen)
	app.timer.Draw(screen)
}

func (app *App) Layout(w, h int) (int, int) {
	return w, h
}
