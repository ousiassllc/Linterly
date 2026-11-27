package reporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSupportedFormat(t *testing.T) {
	for _, f := range []string{"text", "json", "junit", "checkstyle"} {
		assert.True(t, IsSupportedFormat(f), f)
	}
	for _, f := range []string{"", "foo", "JSON", "sarif"} {
		assert.False(t, IsSupportedFormat(f), f)
	}
}
