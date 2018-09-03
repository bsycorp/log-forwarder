package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBackoff(t *testing.T) {
	assert.Equal(t, 0, backoff(0))
	assert.Equal(t, 1, backoff(1))
	assert.Equal(t, 2, backoff(2))
	assert.Equal(t, 4, backoff(3))
	assert.Equal(t, 8, backoff(4))
	assert.Equal(t, 16, backoff(5))
	assert.Equal(t, 32, backoff(6))
	assert.Equal(t, 64, backoff(7))
	assert.Equal(t, 120, backoff(8))
}
