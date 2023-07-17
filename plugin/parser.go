package plugin

import (
	"errors"
	"fmt"
)

// ParseProto parses the proto file and returns a slice of Tables.
func (p *Plugin) ParseProto(provider Provider) ([]fmt.Stringer, error) {
	var tables []fmt.Stringer

	for _, f := range p.req.GetProtoFile() {
		for _, m := range f.GetMessageType() {
			if !isUserMessage(f, m) {
				continue
			}

			var table fmt.Stringer
			switch provider {
			case ProviderPostgres:
				table = newPostgresTable(m)
			case ProviderMysql:
				// todo: implement
			default:
				return nil, ErrUnsupportedProvider
			}

			tables = append(tables, table)
		}
	}

	return tables, nil
}

var ErrUnsupportedProvider = errors.New("unsupported provider")
