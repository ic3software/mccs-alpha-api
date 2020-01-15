package sum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	expected := 10
	actual := Add(5, 5)
	assert.Equal(t, expected, actual)
}
