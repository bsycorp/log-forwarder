package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogBuffer(t *testing.T) {
	b := LogBuffer{}
	b.Append("my dog has fleas")
	assert.Equal(t, b.TotalBytes, 16)
	assert.Equal(t, len(b.Messages), 1)
	assert.False(t, b.NeedsFlush())
	// TODO: test age, probably need clockwork to mock time
}
