package timer

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

const timeFontSize = 32

type timerState int

const (
	stateInitial timerState = iota // before starting
	stateRunning                   // timer running and showing updated value
	statePaused                    // timer running but showing value at pause time
)

type Timer struct {
	startedAt, pausedAt time.Time
	state               timerState

	font     font.Face
	pos      image.Point
	size     image.Point
	timeSize image.Point
}

func New(dimensions image.Rectangle) (*Timer, error) {
	ttf, err := truetype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}

	font := truetype.NewFace(ttf, &truetype.Options{
		Size:    timeFontSize,
		Hinting: font.HintingFull,
	})

	return &Timer{
		font:     font,
		pos:      dimensions.Min,
		size:     dimensions.Size(),
		timeSize: text.BoundString(font, format(time.Duration(0))).Size(),
	}, nil
}

func format(d time.Duration) string {
	return fmt.Sprintf(
		"%d:%02d:%02d.%02d",
		int(math.Floor(d.Hours())),
		int(math.Floor(d.Minutes()))%60,
		int(math.Floor(d.Seconds()))%60,
		(d.Milliseconds()/10)%100,
	)
}

func (timer *Timer) Draw(screen *ebiten.Image) {
	pos := timer.pos.Add(image.Point{
		((timer.size.X - timer.timeSize.X) / 2),
		timer.timeSize.Y + ((timer.size.Y - timer.timeSize.Y) / 2),
	})

	var str string
	switch timer.state {
	case stateInitial:
		// HARDCODED, time size is cached and I don't want to compute this
		pos.X = ((timer.size.X - 19) / 2)
		str = "-"
	case stateRunning:
		str = format(time.Since(timer.startedAt).Round(time.Millisecond))
	case statePaused:
		str = format(timer.pausedAt.Sub(timer.startedAt).Round(time.Millisecond))
	}

	textColor := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	if timer.state == statePaused {
		textColor = color.RGBA{0xDC, 0xAC, 0x26, 0xFF}
	}

	text.Draw(screen, str, timer.font, pos.X, pos.Y, textColor)
}

func (timer *Timer) Toggle() {
	switch timer.state {
	case stateInitial:
		timer.startedAt = time.Now()
		timer.state = stateRunning
	case stateRunning:
		timer.pausedAt = time.Now()
		timer.state = statePaused
	case statePaused:
		timer.startedAt = timer.startedAt.Add(time.Since(timer.pausedAt))
		timer.state = stateRunning
	}
}

func (timer *Timer) Reset() {
	timer.state = stateInitial
}

func (timer *Timer) CanReset() bool {
	return timer.state == statePaused || timer.state == stateInitial
}

func (timer *Timer) IsRunning() bool {
	return timer.state != stateInitial
}

func (timer *Timer) Save() error {
	f, err := os.OpenFile(getSavePath(), os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	pausedAt := timer.pausedAt
	state := statePaused
	switch timer.state {
	case stateInitial:
		pausedAt = time.Time{}
		state = stateInitial
	case stateRunning:
		pausedAt = time.Now()
	case statePaused:
		// NOP
	}

	enc := json.NewEncoder(f)
	return enc.Encode(struct {
		StartedAt, PausedAt time.Time
		State               timerState
	}{
		StartedAt: timer.startedAt,
		PausedAt:  pausedAt,
		State:     state,
	})
}

func (timer *Timer) Load() error {
	f, err := os.Open(getSavePath())
	if err != nil {
		return err
	}
	defer f.Close()

	var s struct {
		StartedAt, PausedAt time.Time
		State               timerState
	}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&s); err != nil {
		return err
	}

	timer.startedAt = s.StartedAt
	timer.pausedAt = s.PausedAt
	timer.state = s.State

	return nil
}

func getSavePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = "./"
	}

	return filepath.Join(dir, "ivan.timer.state.json")
}
