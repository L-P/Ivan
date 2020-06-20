package main

import "fmt"

// Version holds the compile-time version string of Ivan.
// nolint:gochecknoglobals
var Version = "unknown"

func main() {
	fmt.Printf("ivan %s\n", Version)
}
