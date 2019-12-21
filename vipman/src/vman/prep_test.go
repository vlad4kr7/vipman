package vman

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNameMatch(t *testing.T) {
	assert.True(t, Match("", ""))
	assert.True(t, Match("a*", "a"))
	assert.True(t, Match("a*", "abcd"))
	assert.True(t, Match("abcd*", "abcd"))
	assert.True(t, Match("abcd*", "abcde"))
	assert.True(t, !Match("abcde*", "abcd"))
}
