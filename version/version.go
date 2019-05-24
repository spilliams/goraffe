package version

import "fmt"

// set by the linker
var versionNumber = "Unknown"
var gitHash = "Unknown"
var buildTime = "Unknown"

// Info returns the current version info
func Info() string {
	return fmt.Sprintf("version %s\nbuilt %s\nhash %s", versionNumber, buildTime, gitHash)
}
