package main

import (
	"errors"
	"fmt"
	inputviewer "ivan/input-viewer"
	"ivan/timer"
	"ivan/tracker"
	"log"
	"time"

	"github.com/bep/debounce"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const configPath = "assets/config.json"

var errCloseApp = errors.New("user requested app close")

type App struct {
	tracker     *tracker.Tracker
	timer       *timer.Timer
	inputViewer *inputviewer.InputViewer
	config      config
	lastSave    time.Time

	saveDebounce func(func())
}

func NewApp() (*App, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load initial config: %w", err)
	}

	size := config.windowSize()
	ebiten.SetWindowSize(size.X, size.Y)
	ebiten.SetWindowPosition(1920-size.X, 0)

	timer, err := timer.New(config.Dimensions.Timer)
	if err != nil {
		return nil, err
	}

	tracker, err := tracker.New(
		config.Dimensions.ItemTracker,
		config.Dimensions.HintTracker,
		config.Items,
		config.ZoneItemMap,
		config.Locations,
		config.Binds,
		config.AlwaysHints,
		config.DungeonInputMedallionOrder,
		config.DungeonInputDungeonKP,
	)
	if err != nil {
		return nil, err
	}

	if err := tracker.Load(); err != nil {
		log.Printf("error: %s", err)
	}

	if err := timer.Load(); err != nil {
		log.Printf("error: %s", err)
	}

	return &App{
		tracker:      tracker,
		timer:        timer,
		inputViewer:  nil, // initialized on first frame to ensure we have a gamepad
		config:       config,
		saveDebounce: debounce.New(1 * time.Second),
		lastSave:     time.Now(),
	}, nil
}

// nolint: funlen
func (app *App) Update() error {
	_, wheel := ebiten.Wheel()
	var shouldSave bool

	if app.inputViewer == nil && app.config.InputViewer.Enabled {
		app.inputViewer = inputviewer.NewInputViewer(app.config.InputViewer)
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		if !app.timer.IsRunning() && app.tracker.IsIdle() {
			return errCloseApp
		}
		app.tracker.Cancel()

	case inpututil.IsKeyJustPressed(ebiten.KeyHome):
		if !app.timer.IsRunning() {
			config, err := loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("unable to load config: %w", err)
			}
			app.config = config
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
		app.tracker.Submit()
		shouldSave = true

	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if app.tracker.EatInput() {
			app.tracker.Input([]rune(" "))
			break
		}
		app.timer.Toggle()

	case inpututil.IsKeyJustPressed(ebiten.KeyDelete):
		if app.timer.CanReset() {
			app.timer.Reset()
			app.tracker.Reset(app.config.Items, app.config.ZoneItemMap)
			shouldSave = true
		}

	case inpututil.IsKeyJustPressed(ebiten.KeyEnd):
		shouldSave = true

	case inpututil.IsKeyJustPressed(ebiten.KeyBackspace):
		app.tracker.Backspace()

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		app.tracker.ClickLeft(ebiten.CursorPosition())
		shouldSave = true

	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight):
		app.tracker.ClickRight(ebiten.CursorPosition())
		shouldSave = true

	case wheel != 0:
		x, y := ebiten.CursorPosition()
		app.tracker.Wheel(x, y, wheel > 0)
		shouldSave = true

	default:
		input := ebiten.InputChars()
		if len(input) > 0 {
			app.tracker.Input(input)
			shouldSave = true
		}
	}

	if shouldSave || time.Since(app.lastSave) > (10*time.Second) {
		app.lastSave = time.Now()
		app.saveDebounce(func() {
			log.Print("info: saving")
			if err := app.tracker.Save(); err != nil {
				log.Printf("error: unable to write tracker save: %s", err)
			}
			if err := app.timer.Save(); err != nil {
				log.Printf("error: unable to write timer save: %s", err)
			}
		})
	}

	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	app.tracker.Draw(screen)
	app.timer.Draw(screen)
	app.inputViewer.Draw(screen)
}

func (app *App) Layout(w, h int) (int, int) {
	return w, h
}
