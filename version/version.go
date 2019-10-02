package version

import (
	"fmt"
)

var (
	// GitCommit that was compiled. This will be filled in by the compiler
	GitCommit string

	// Version number that is being run at the moment
	Version = "0.3.0"
)

// FullVersion returns the B2 version and the commit hash
func FullVersion() string {
	return fmt.Sprintf("B2 v%s (%s)", Version, GitCommit)
}
