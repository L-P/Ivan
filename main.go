package main

import (
	"log"

	_ "image/png"

	"github.com/hajimehoshi/ebiten"
)

// Version holds the compile-time version string of Ivan.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	log.Printf("ivan %s\n", Version)

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("Ivan")
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowDecorated(false)

	ivan, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(ivan); err != nil && err != errCloseApp {
		log.Fatal(err)
	}

	log.Println("ivan closed")
}
