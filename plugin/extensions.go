package plugin

type ExtensionSet map[Extension]bool

type Extension string

func (e Extension) String() string {
	return string(e)
}

func (e ExtensionSet) Enable(extensions ...Extension) {
	for _, v := range extensions {
		e[v] = true
	}
}

var (
	// postgres extensions uuid-ossp
	ExtensionUUID = Extension("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
)
