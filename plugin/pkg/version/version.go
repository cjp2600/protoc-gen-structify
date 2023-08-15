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

func GetPluginVersion() string {
	return fmt.Sprintf("(version=%s, branch=%s, revision=%s), build: %s", Version, Branch, Revision, VersionBuildContext())
}

func VersionBuildContext() string {
	return fmt.Sprintf("(go=%s, date=%s)", GoVersion, BuildDate)
}
