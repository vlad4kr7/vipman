package vman

import (
	"github.com/stretchr/testify/assert"
	"math"
	"strings"
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

func TestNameSplit(t *testing.T) {
	assert.Equal(t, 1, len(strings.Split("a", ":")))
	assert.Equal(t, 2, len(strings.Split("a:", ":")))
	assert.Equal(t, 2, len(strings.Split("a:b", ":")))
}

func TestMod(t *testing.T) {
	assert.Equal(t, 0, int(math.Mod(4, 4)))
	assert.Equal(t, 1, int(math.Mod(5, 4)))
}

func TestStrCmp(t *testing.T) {
	assert.Equal(t, 6, compMax([]string{"1.2.3.4", "1.2.3.5", "1.2.3.5"}))
	assert.Equal(t, 4, compMax([]string{"1.2.2.4", "1.2.3.5", "1.2.3.5"}))
	assert.Equal(t, 5, compMax([]string{"10.0.0.4", "10.0.3.5", "10.0.30.5"}))
	assert.Equal(t, 0, compMax([]string{"10.0.0.1", "10.0.0.1"}))
	assert.Equal(t, 5, compMax([]string{"10.0.0.40", "10.0.3.5", "10.0.30.5"}))
	assert.Equal(t, 3, compMax([]string{"10.0.20.4", "10.0.30.5", "10.5.30.5"}))
}
