package main

import (
	"github.com/cjp2600/protoc-gen-structify/plugin"
)

func main() {
	// Run the plugin
	// This will read the request from stdin, and write the response to stdout
	plugin.NewPlugin().Run()
}
