package main

import (
	"os"

	"github.com/Sho2010/dup-finder/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
