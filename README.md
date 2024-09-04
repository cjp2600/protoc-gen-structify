# protoc-gen-structify

[![Go Report Card](https://goreportcard.com/badge/github.com/cjp2600/protoc-gen-structify)](https://goreportcard.com/report/github.com/cjp2600/protoc-gen-structify)

`protoc-gen-structify` is a `protoc` plugin designed to generate structured Go data types from your Protocol Buffers (protobuf) definitions. It provides an easy way to create Go structs that match your protobuf messages, enhancing the integration between protobuf and Go.

## Features

- Automatically generates Go structs based on protobuf messages.
- Supports both simple and complex protobuf types.
- Maintains field names, types, and tags consistent with the protobuf definitions.
- Example usage available in the `examples` directory.

## Installation

To install `protoc-gen-structify`, run the following command:

```bash
go install github.com/cjp2600/protoc-gen-structify@latest
```
Make sure that your GOPATH/bin is added to your PATH environment variable so that protoc can find the plugin.

## Usage
To use protoc-gen-structify with protoc, run the following command in your project directory: