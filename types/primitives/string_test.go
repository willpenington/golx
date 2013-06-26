package primitives

import (
  "golx/types"
  "testing"
)

func TestStringIsObject(t *testing.T) {
  s := NewString("hello")
  _, success := s.(types.Object)

  if !success {
    t.Fail()
  }
}

