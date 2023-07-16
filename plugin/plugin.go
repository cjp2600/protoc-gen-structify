package plugin

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	plugingo "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Plugin struct {
	req *plugingo.CodeGeneratorRequest
	res *plugingo.CodeGeneratorResponse

	Param    map[string]string
	genFiles []*plugingo.CodeGeneratorResponse_File
	pathType pathType // How to generate output filenames.
}

func NewPlugin() *Plugin {
	return &Plugin{
		req: &plugingo.CodeGeneratorRequest{},
		res: &plugingo.CodeGeneratorResponse{},
	}
}

func (p *Plugin) commandLineParameters(parameter string) {
	p.Param = make(map[string]string)
	for _, d := range strings.Split(parameter, ",") {
		if i := strings.Index(d, "="); i < 0 {
			p.Param[d] = ""
		} else {
			p.Param[d[0:i]] = d[i+1:]
		}
	}
	for k, v := range p.Param {
		switch k {
		case "paths":
			switch v {
			case "import":
				p.pathType = pathTypeImport
			case "source_relative":
				p.pathType = pathTypeSourceRelative
			default:
				p.Error(fmt.Errorf(`Unknown path type %q: want "import" or "source_relative".`, v))
			}
		}
	}
}

func (p *Plugin) Run() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		p.Error(err)
	}

	if err := proto.Unmarshal(data, p.req); err != nil {
		p.Error(err)
	}
	p.commandLineParameters(p.req.GetParameter())

	p.fill()

	tables, err := p.ParseProto()
	if err != nil {
		p.Error(err)
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

	if err := goFmt(p.res); err != nil {
		p.Error(err)
	}

	data, err = proto.Marshal(p.res)
	if err != nil {
		p.Error(err)
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		p.Error(err)
	}
}

func (p *Plugin) fill() {
	p.genFiles = make([]*plugingo.CodeGeneratorResponse_File, 0, len(p.res.GetFile()))
	for _, file := range p.res.GetFile() {
		p.genFiles = append(p.genFiles, file)
	}
}

func (p *Plugin) Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-structify: error:", s)
	os.Exit(1)
}

type pathType int

const (
	pathTypeImport pathType = iota
	pathTypeSourceRelative
)

const generatedFilePostfix = ".db.go"
