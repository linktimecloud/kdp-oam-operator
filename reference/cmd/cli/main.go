package main

import (
	"kdp-oam-operator/reference/cli"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	command := cli.NewCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
