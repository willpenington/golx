/*
GoLX Type system

In order to avoid depending too heavily on Go's reflection and to confine the
as much of the reflection magic as possible to one place GoLX has it's own type
system for high level operations. This also unifies behaviour like fixture
profiles into a classic inheritance scheme rather than having several ad hoc
systems.

The system is currently very simple and does not yet support inheritance or
custom methods. All methods are found using Go reflection.
*/
package types

import "reflect"

type Type interface {
	Name() string

  // Embed the Object interface so all types are objects
  Type() Type
  Method(name string) Function
}

type Object interface {
	Type() Type
  Method(name string) Function
}

/* Simple type */
type SimpleType struct {
  TypeName string
}

func (st SimpleType) Name() string {
  return st.TypeName
}

func (SimpleType) Type() Type {
  return MetaType()
}

func (st SimpleType) Method(name string) Function {
  return Method(st, name)
}

/* Type of Type */
/* All types should also be objects that give their type as the meta type */
func MetaType() Type {
  return SimpleType{TypeName: "Type"}
}

/* Method calls */

// Return the go method as a Function object if it uses Objects
func Method(o Object, name string) Function {
	rval := reflect.ValueOf(o)
	rmethod := rval.MethodByName(name)

	if rmethod.IsValid() {
		f, err := NewFunction(rmethod.Interface())
		if err != nil {
			return f
		}
	}

	return nil
}
