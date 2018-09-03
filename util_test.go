package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListContains(t *testing.T) {
	a := []string{"foo", "bar"}
	assert.True(t, ListContains(a, "foo"))
	assert.False(t, ListContains(a, "baz"))
}

func TestListIntersect(t *testing.T) {
	a := []string{"foo", "baz"}
	b := []string{"a", "foo", "b", "bim"}
	c := []string{}
	assert.Equal(t, []string{"foo"}, ListIntersect(a, b))
	assert.Equal(t, c, ListIntersect(a, c))
}

func TestListSubtract(t *testing.T) {
	a := []string{"foo", "bar", "baz"}
	b := []string{"bar"}
	assert.Equal(t, []string{"foo", "baz"}, ListSubtract(a, b))
	assert.Equal(t, []string{}, ListSubtract(b, a))
}
