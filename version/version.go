package version

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	version = `dev`
	built   = ``
	commit  = ``
	info    Info
)

type Info struct {
	Version    string
	BuildDate  time.Time
	CommitHash string
}

func init() {
	info = Info{
		Version:    version,
		CommitHash: commit,
	}
	info.BuildDate, _ = parseBuildTime()
}

//goland:noinspection GoBoolExpressions
func (Info Info) String() string {
	var sb strings.Builder
	sb.WriteString(Info.Version)
	if Info.CommitHash != `` {
		sb.WriteString(`-`)
		sb.WriteString(Info.CommitHash)
	}
	if !Info.BuildDate.IsZero() {
		sb.WriteString(`-`)
		sb.WriteString(Info.BuildDate.Format(time.RFC3339))
	}

	sb.WriteString(` (`)
	sb.WriteString(runtime.Version())
	sb.WriteString(`)`)
	return sb.String()
}

//goland:noinspection GoBoolExpressions
func (Info Info) LongString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Version:     %s\n", info.Version))
	sb.WriteString(fmt.Sprintf("Go Version:  %s\n", runtime.Version()))
	if info.CommitHash != `` {
		sb.WriteString(fmt.Sprintf("Git Commit:  %s\n", info.CommitHash))
	}
	if !Info.BuildDate.IsZero() {
		sb.WriteString(fmt.Sprintf("Built:       %s\n", Info.BuildDate.Format(time.RFC3339)))
	}
	sb.WriteString(fmt.Sprintf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH))
	return sb.String()
}

func Get() Info {
	return info
}

func parseBuildTime() (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, built); err == nil {
		return t, true
	}
	if builtInt, err := strconv.ParseInt(built, 10, 64); err == nil {
		return time.Unix(builtInt, 0).UTC(), true
	}
	return time.Time{}, false
}
