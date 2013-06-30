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

// Type 
type Type interface {
	Name() string
  Method(name string) Method
  Methods() []string
}

type Object interface {
	Type() Type
	ObjectMethod(name string) Function
}

type Method interface {
  // Returns a closure of the method over an object
  Apply(Object) (Function, error)
}
