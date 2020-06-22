package main

import (
	"log"
	"runtime"

	_ "image/png"

	"github.com/hajimehoshi/ebiten"
)

// Version holds the compile-time version string of Ivan.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	log.Printf("ivan %s\n", Version)

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowPosition(1920-width, 0)
	ebiten.SetWindowTitle("Ivan")
	ebiten.SetRunnableOnUnfocused(true)

	// Can't move an undecorated window on Windows… Such good UX.
	if runtime.GOOS != "windows" {
		ebiten.SetWindowDecorated(false)
	}

	ivan, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(ivan); err != nil && err != errCloseApp {
		log.Fatal(err)
	}
}
