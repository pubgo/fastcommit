package utils

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestIsDirty(t *testing.T) {
	assert.NoError(t, IsDirty().GetErr())
}
