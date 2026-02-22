package buildinfo

import (
	"fmt"

	"icekalt.dev/money-tracker/internal/devmode"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

func String() string {
	s := fmt.Sprintf("money-tracker %s (commit: %s, built: %s, go: %s)", Version, Commit, BuildDate, GoVersion)
	if devmode.Enabled {
		s += " [DEV BUILD]"
	}
	return s
}
