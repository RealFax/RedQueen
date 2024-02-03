package version

import "strings"

const (
	Major = "0"
	Minor = "8"
	Patch = "1"
)

var (
	BuildTime    string
	BuildVersion string
)

var str string

func init() {
	b := strings.Builder{}
	b.WriteString("v" + Major + "." + Minor + "." + Patch)

	if BuildVersion != "" {
		b.WriteString(", build version: " + BuildVersion)
	}

	if BuildTime != "" {
		b.WriteString(", time: " + BuildTime)
	}

	str = b.String()
}

func String() string {
	return str
}
