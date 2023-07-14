package plugin

import (
	structify "github.com/cjp2600/structify/plugin/options"
	"github.com/golang/protobuf/proto"
	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/types/descriptorpb"
)

func getMessageOptions(d *descriptorpb.DescriptorProto) *structify.StructifyMessageOptions {
	opts := d.GetOptions()
	if opts != nil {
		ext, _ := proto.GetExtension(opts, structify.E_Opts)
		if ext != nil {
			customOpts, ok := ext.(*structify.StructifyMessageOptions)
			if ok {
				return customOpts
			}
		}
	}
	return nil
}

func isUserMessage(f *descriptorpb.FileDescriptorProto, m *descriptorpb.DescriptorProto) bool {
	if f.GetPackage() == "google.protobuf" || f.GetPackage() == "structify" {
		return false
	}

	return true
}

func sToCml(name string) string {
	return strcase.UpperCamelCase(name)
}

func prepareType(s string) string {
	switch s {
	case "TYPE_INT32":
		return "int32"
	case "TYPE_STRING":
		return "string"
	}
	return s
}
