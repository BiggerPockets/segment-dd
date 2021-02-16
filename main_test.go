package main

import (
  "testing"
)

import (
  "github.com/stretchr/testify/assert"
)

func TestFormatEventName(t *testing.T) {
  assert.Equal(t, "viewed_dashboard", formatEventName("Viewed Dashboard"))
}

func TestValidEvent(t *testing.T) {
  loadConfig();

  assert.True(t, validEvent("Viewed Dashboard"))
  assert.False(t, validEvent("Consumed Ice Cream"))
}
