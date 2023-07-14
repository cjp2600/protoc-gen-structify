package plugin

import (
	"io"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"

	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Plugin struct {
	req *plugingo.CodeGeneratorRequest
	res *plugingo.CodeGeneratorResponse
}

func NewPlugin() *Plugin {
	return &Plugin{
		req: &plugingo.CodeGeneratorRequest{},
		res: &plugingo.CodeGeneratorResponse{},
	}
}

func (p *Plugin) Run() error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(data, p.req); err != nil {
		return err
	}

	tables, err := p.ParseProto()
	if err != nil {
		return err
	}

	content := &strings.Builder{}
	content.WriteString("package main\n\n")

	for _, t := range tables {
		content.WriteString(t.String())
		content.WriteString("\n")
	}

	p.res.File = append(p.res.File, &plugingo.CodeGeneratorResponse_File{
		Name:    proto.String("pro" + generatedFilePostfix),
		Content: proto.String(content.String()),
	})

	data, err = proto.Marshal(p.res)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)
	return err
}

const generatedFilePostfix = ".db.go"
