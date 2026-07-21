// Package buildinfo contains release metadata injected by the build.
package buildinfo

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
)

func String() string {
	if Commit == "" || Commit == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s (%s)", Version, Commit)
}
