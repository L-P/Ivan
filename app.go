package main

import (
	"errors"
	"image"
	"ivan/timer"
	"ivan/tracker"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/inpututil"
)

const (
	width  = tracker.Width
	height = tracker.Height + timer.Height
)

var errCloseApp = errors.New("user requested app close")

type App struct {
	background *ebiten.Image
	tracker    *tracker.Tracker
	timer      *timer.Timer
}

func NewApp() (*App, error) {
	background, _, err := ebitenutil.NewImageFromFile("assets/background.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	timer, err := timer.New(image.Point{0, tracker.Height})
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
		timer:      timer,
	}, nil
}

func (app *App) Update(screen *ebiten.Image) error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		if !app.timer.IsRunning() {
			return errCloseApp
		}
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		app.timer.Toggle()
	case inpututil.IsKeyJustPressed(ebiten.KeyDelete):
		app.timer.Reset()
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		app.tracker.Upgrade(ebiten.CursorPosition())
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		app.tracker.Downgrade(ebiten.CursorPosition())
	default:
		return nil
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	if err := screen.DrawImage(app.background, nil); err != nil {
		log.Fatal(err)
	}

	app.tracker.Draw(screen)
	app.timer.Draw(screen)
}

func (app *App) Layout(w, h int) (int, int) {
	return width, height
}
