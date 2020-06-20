package main

import (
	"fmt"
	"log"

	_ "image/png"

	"github.com/hajimehoshi/ebiten"
)

// Version holds the compile-time version string of Ivan.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	fmt.Printf("ivan %s\n", Version)

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("Ivan")

	ivan, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(ivan); err != nil {
		log.Fatal(err)
	}
}
