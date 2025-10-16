package main

import (
	"os"

	"olive-clone-assistant-v2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}