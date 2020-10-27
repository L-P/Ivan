package timer

import (
	"fmt"
	"image"
	"image/color"
	"math"
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
