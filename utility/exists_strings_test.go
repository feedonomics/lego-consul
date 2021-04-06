package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistsStrings(t *testing.T) {
	v := []string{`test1`, `test2`}
	assert.True(t, ExistsStrings(v, `test1`))
	assert.True(t, ExistsStrings(v, `test2`))
	assert.False(t, ExistsStrings(v, `test3`))
}
