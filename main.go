package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	_ "image/png"

	"github.com/hajimehoshi/ebiten"
)

// Version holds the compile-time version string of Ivan.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	log.Printf("ivan %s\n", Version)

	chdirToExecutableDir()

	ebiten.SetWindowTitle("Ivan")
	ebiten.SetWindowSize(width, height)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowResizable(true)

	ivan, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(ivan); err != nil && err != errCloseApp {
		log.Fatal(err)
	}
}

func chdirToExecutableDir() {
	exec, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	solved, err := filepath.EvalSymlinks(exec)
	if err != nil {
		log.Printf("warning: %s", err)
		solved = exec
	}

	if err := os.Chdir(path.Dir(solved)); err != nil {
		log.Fatal(err)
	}
}
