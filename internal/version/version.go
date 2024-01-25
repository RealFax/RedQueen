package version

import "fmt"

const (
	Major = "0"
	Minor = "7"
	Patch = "1"
)

var (
	BuildTime    string
	BuildVersion string
)

var v = fmt.Sprintf("v%s.%s.%s, build version: %s, time: %s", Major, Minor, Patch, BuildVersion, BuildTime)

func String() string {
	return v
}
