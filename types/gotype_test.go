package types

import "testing"

func TestGoObjectsForTheSameValueAreEqual(t *testing.T) {
	v := "asdf"
	a := NewGoObject(v)
	b := NewGoObject(v)

	if a != b {
		t.Fail()
	}
}

func TestGoObjectsForDifferentValuesAreNotEqual(t *testing.T) {
	a := NewGoObject("asdf")
	b := NewGoObject("qewr")

	if a == b {
		t.Fail()
	}
}

func TestGoTypesOfValuesWithTheSameTypeAreEqual(t *testing.T) {
	a := NewGoObject("asdf")
	b := NewGoObject("qwer")

	if a.Type() != b.Type() {
		t.Fail()
	}
}

func TestGoTypesOfValuesOfDifferentTypeAreDifferent(t *testing.T) {
	a := NewGoObject("asdf")
	b := NewGoObject(123)

	if a.Type() == b.Type() {
		t.Fail()
	}
}

func TestGoTypeOfFunctionIsFunction(t *testing.T) {
	f := func(a int) int { return a + 1 }
	a := NewGoObject(f)

	_, isFunc := a.(Function)

	if !isFunc {
		t.Fail()
	}
}

type a struct{}

func (a) Blah(c int) int {
	return c + 1
}

func TestGoObjectReturnsWrappedGoMethods(t *testing.T) {
	obj := NewGoObject(a{})
	blah := obj.Method("Blah")

	ret, err := blah.Call([]Object{NewGoObject(3)})

	if err != nil || (len(ret) == 1 && ret[0] != NewGoObject(4)) {
		t.Fail()
	}
}
