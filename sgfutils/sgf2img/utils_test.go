package sgf2img

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandCoordinates(t *testing.T) {
	t.Parallel()

	for _, data := range []struct {
		str      string
		expected []string
	}{
		{"dd:de", []string{"dd", "de"}},
		{"dd:df", []string{"dd", "de", "df"}},
		{"dd:ee", []string{"dd", "de", "ed", "ee"}},
	} {
		coords, err := expandCoordinatesRange(data.str, 19)
		assert.Nil(t, err)
		assert.Equal(t, data.expected, coords)
	}
}
