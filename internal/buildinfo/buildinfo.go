package buildinfo

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

func String() string {
	return fmt.Sprintf("money-tracker %s (commit: %s, built: %s, go: %s)", Version, Commit, BuildDate, GoVersion)
}
