package cli

import (
	"encoding/xml"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCheck_XMLFormats(t *testing.T) {
	cases := []struct {
		name     string
		lines    int
		wantCode int // 0 = 正常終了
	}{
		{"violation", 20, ExitViolation},
		{"warn_only", 8, 0},
		{"pass", 1, 0},
	}
	for _, fmtName := range []string{"junit", "checkstyle"} {
		for _, tc := range cases {
			t.Run(fmtName+"/"+tc.name, func(t *testing.T) {
				output, err := runCheckWithFormat(t, fmtName, tc.lines)

				if tc.wantCode == 0 {
					require.NoError(t, err)
				} else {
					var exitErr *ExitError
					require.True(t, errors.As(err, &exitErr))
					assert.Equal(t, tc.wantCode, exitErr.Code)
				}

				var doc struct{ XMLName xml.Name }
				require.NoError(t, xml.Unmarshal([]byte(output), &doc), "output should be valid XML: %s", output)
				wantRoot := map[string]string{"junit": "testsuites", "checkstyle": "checkstyle"}[fmtName]
				assert.Equal(t, wantRoot, doc.XMLName.Local)
			})
		}
	}
}
