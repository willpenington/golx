package types

import (
  "reflect"
  "errors"
)

/*
Wraps Go values in a GoLX Objects
*/
type GoObject reflect.Value

/*
Wraps Go types in GoLX Types
*/
type GoType struct {
  refType reflect.Type
}

/*
Wraps Go functions in GoLX functions
*/
type GoFunc reflect.Value

/*
Get a Go object as it's corresponding GoLX type
*/
func NewGoObject(val interface {}) Object {
  if reflect.ValueOf(val).Kind() == reflect.Func {
    // Functions have a special type
    return GoFunc(reflect.ValueOf(val))
  } else {
    return GoObject(reflect.ValueOf(val))
  }
}

/*
GoLX Type for an object implemented in Go
*/
func (obj GoObject) Type() Type {
  return GetGoType(reflect.TypeOf(obj))
}

/*
Retrieve methods on Go objects. Uses reflection to call 
*/
func (obj GoObject) Method(name string) Function {
  val := reflect.ValueOf(obj)
  method := val.MethodByName(name)
  if method.IsValid() {
    return method.Interface().(GoFunc)
  } else {
    return Function(nil)
  }
}

/*
Name of the GoLX type for types implemented in Go
*/
func (t GoType) Name() string {
  return "go:" + t.refType.Name()
}

/*

*/
func GetGoType(t reflect.Type) Type {
  if t.Kind() == reflect.Func {
    // The error will only be returned t isn't a function, but here it is garunteed to be
    ftype, _ := GetFuncType(t)
    return ftype
  } else {
    return GoType{refType: t}
  }
}

func GetFuncType(t reflect.Type) (FunctionType, error) {
  if t.Kind() != reflect.Func {
    return FunctionType{}, errors.New("Argument must be a go function type")
  }

  args := make([]Type, t.NumIn())
  rets := make([]Type, t.NumOut())

  for i := range args {
    args[i] = GetGoType(t.In(i))
  }

  for i := range rets {
    rets[i] = GetGoType(t.Out(i))
  }

  return NewFunctionType(args, rets), nil
}

func (f GoFunc) Method(name string) Function {
  return GoObject(f).Method(name)
}

func (f GoFunc) FunctionType() FunctionType {
  ft, err := GetFuncType(reflect.Value(f).Type())
  if err != nil {
    panic("Only functions can be cast to GoFunc")
  }
  return ft
}

func (f GoFunc) Type() Type {
  return Type(f.FunctionType())
}


func (f GoFunc) Call(args []Object) ([]Object, error) {
  fType := f.FunctionType()
  fVal := reflect.Value(f)
  argVals := make([]reflect.Value, len(args))

  if len(args) != len(fType.Args()) {
    return []Object{}, errors.New("Incorrect number of arguments")
  }

  for i, arg := range args {
    if arg.Type() != fType.Args()[i] {
      return []Object{}, errors.New("Incorrect argument type for arg " + string(i))
    }

    argType, argIsGoType := arg.(reflect.Type)

    if argIsGoType {
      argVals[i] = reflect.ValueOf(args[i]).Convert(argType)
    } else {
      argVals[i] = reflect.ValueOf(args[i])
    }
  }

  retVals := fVal.Call(argVals)

  rets := make([]Object, len(retVals))

  for i, retVal := range retVals {
    rets[i] = retVal.Interface().(Object)
  }

  return rets, nil
}

