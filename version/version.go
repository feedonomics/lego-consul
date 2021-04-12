package version

import "strings"

const version = "0.0.9"

var (
	commit = ""
	date   = ""
	info   Info
)

// Info defines version details
type Info struct {
	Version    string `json:"version"`
	BuildDate  string `json:"build_date"`
	CommitHash string `json:"commit_hash"`
}

// GetAsString returns the string representation of the version
func String() string {
	var sb strings.Builder
	sb.WriteString(info.Version)
	if len(info.CommitHash) > 0 {
		sb.WriteString("-")
		sb.WriteString(info.CommitHash)
	}
	if len(info.BuildDate) > 0 {
		sb.WriteString("-")
		sb.WriteString(info.BuildDate)
	}
	return sb.String()
}

func init() {
	info = Info{
		Version:    version,
		CommitHash: commit,
		BuildDate:  date,
	}
}

// Get returns the Info struct
func Get() Info {
	return info
}
