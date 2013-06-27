package types

/* A simple type that stores a name to return */

type SimpleType struct {
	name string
}

func GetSimpleType(name string) SimpleType {
	var st SimpleType
	st.name = name
	return st
}

func (st SimpleType) Name() string {
	return st.name
}

func (SimpleType) Method(name string) Function {
  return nil
}
