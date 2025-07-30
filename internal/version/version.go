package version

import (
	"fmt"
	"runtime"
)

// Build information set via ldflags
var (
	// Version is the current version of the application
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// Date is the build date
	Date = "unknown"

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// Info holds version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("Pepo %s (commit: %s, built: %s, go: %s)",
		i.Version, i.Commit, i.Date, i.GoVersion)
}

// Short returns a short version string
func (i Info) Short() string {
	return fmt.Sprintf("Pepo %s", i.Version)
}

// Print prints version information to stdout
func Print() {
	fmt.Println(Get().String())
}
