package main

import (
	"errors"
	_ "image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

// Version holds the compile-time version string of Ivan.
var Version = "unknown"

func main() {
	log.Printf("ivan %s\n", Version)

	chdirToExecutableDir()

	ebiten.SetWindowTitle("Ivan")
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Undecorated windows can't be moved under Windows or Darwin.
	if runtime.GOOS == "linux" {
		ebiten.SetWindowDecorated(false)
	}

	ivan, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(ivan); err != nil && !errors.Is(err, errCloseApp) {
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
