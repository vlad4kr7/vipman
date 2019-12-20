package vman

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNameMatch(t *testing.T) {
	assert.True(t, Match("", ""))
	assert.True(t, Match("a*", "a"))
	assert.True(t, Match("a*", "abcd"))
	assert.True(t, Match("abcd*", "abcd"))
	assert.True(t, !Match("abcde*", "abcd"))
}