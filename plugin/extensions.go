package plugin

// ExtensionSet is a set of extensions.
type ExtensionSet map[Extension]bool

// Extension is a type for how to generate extension statements.
// Example: CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
type Extension string

// String returns a string representation of the ExtensionSet.
func (e Extension) String() string {
	return string(e)
}

// Enable enables extensions in the ExtensionSet.
func (e ExtensionSet) Enable(extensions ...Extension) {
	for _, v := range extensions {
		e[v] = true
	}
}

var (
	// ExtensionUUID postgres extensions uuid-ossp
	ExtensionUUID = Extension("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
)
