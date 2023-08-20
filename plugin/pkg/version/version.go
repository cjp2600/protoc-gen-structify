package version

import (
	"fmt"
	"runtime"
)

var (
	Version   string
	Revision  string
	Branch    string
	BuildDate string
	GoVersion = runtime.Version()
)

// GetPluginVersion returns the version of the plugin.
func GetPluginVersion() string {
	return fmt.Sprintf("(version=%s, branch=%s, revision=%s), build: %s", Version, Branch, Revision, VersionBuildContext())
}

// VersionBuildContext returns the build context.
func VersionBuildContext() string {
	return fmt.Sprintf("(go=%s, date=%s)", GoVersion, BuildDate)
}
