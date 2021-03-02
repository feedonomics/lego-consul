package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSANs(t *testing.T) {
	Domains := ParseSANs([]string{
		`mydomain2.com, mydomain3.com`,
		`mydomain4.com`,
	})

	assert.Equal(t, []string{
		`mydomain2.com`,
		`mydomain3.com`,
		`mydomain4.com`,
	}, Domains)
}
