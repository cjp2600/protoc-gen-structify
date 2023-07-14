package main

import (
	"log"

	"github.com/cjp2600/structify/plugin"
)

func main() {
	p := plugin.NewPlugin()

	if err := p.Run(); err != nil {
		log.Fatalf("Failed to run plugin: %v", err)
	}
}
