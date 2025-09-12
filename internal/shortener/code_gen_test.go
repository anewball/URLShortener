package shortener

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNanoID(t *testing.T) {
	nanoID := NewNanoID("abc")

	assert.NotNil(t, nanoID)

	generated, err := nanoID.Generate(5)
	assert.NoError(t, err)
	assert.Len(t, generated, 5)
}
