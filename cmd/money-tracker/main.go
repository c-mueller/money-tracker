package main

import (
	"os"

	"icekalt.dev/money-tracker/cmd/money-tracker/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
