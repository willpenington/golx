package primitives

import "golx/types"

stringType := types.SimpleType{Name: String}

type String string

func (String) Type() Type {
  return StringType
}

func NewString(val string) String {
  return String(val)
}

func (s String) String() string {
  return string(s)
}
