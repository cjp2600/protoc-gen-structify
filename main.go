package main

import (
	"github.com/cjp2600/structify/plugin"
	"github.com/gogo/protobuf/vanity/command"
)

const generatedFilePostfix = ".db.go"

func main() {
	// Run the command from gogo/protobuf/vanity/command
	command.Write(command.GeneratePlugin(
		command.Read(),
		plugin.NewStructifyPlugin(),
		generatedFilePostfix,
	))
}
