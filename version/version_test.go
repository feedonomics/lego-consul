package version

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVersionShort(t *testing.T) {
	info.BuildDate = time.Date(2021, time.November, 12, 14, 34, 17, 0, time.UTC)
	info.Version = `v0.0.1`
	info.CommitHash = `abcd1234`

	expectedShort := fmt.Sprintf(`v0.0.1-abcd1234-2021-11-12T14:34:17Z (%s)`, runtime.Version())
	assert.Equal(t, expectedShort, info.String())
}

func TestVersionLong(t *testing.T) {
	info.BuildDate = time.Date(2021, time.November, 12, 14, 34, 17, 0, time.UTC)
	info.Version = `v0.0.1`
	info.CommitHash = `abcd1234`

	expectedLong := fmt.Sprintf("Version:     v0.0.1\nGo Version:  %s\nGit Commit:  abcd1234\nBuilt:       2021-11-12T14:34:17Z\nOS/Arch:     %s/%s\n",
		runtime.Version(), runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, expectedLong, Get().LongString())
}

func TestParseBuiltTime(t *testing.T) {
	expectedTime := time.Date(2021, time.October, 13, 17, 38, 12, 0, time.UTC)

	built = expectedTime.Format(time.RFC3339)
	actualTime, ok := parseBuildTime()
	assert.True(t, ok)
	assert.True(t, actualTime.Equal(expectedTime))

	built = strconv.FormatInt(expectedTime.Unix(), 10)
	actualTime, ok = parseBuildTime()
	assert.True(t, ok)
	assert.True(t, actualTime.Equal(expectedTime))
}
