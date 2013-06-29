package types

import (
  "reflect"
  "errors"
)

/*
Wraps Go values in a GoLX Objects
*/
type GoObject struct {
  val interface {}
}

/*
Wraps Go types in GoLX Types
*/
type GoType struct {
  refType reflect.Type
}

/*
Wraps Go functions in GoLX functions
*/
type GoFunc struct {
  val interface {}
}

/*
Wraps Go methods as GoLX methods
*/
type GoMethod reflect.Method

/*
Get a Go object as it's corresponding GoLX type
*/
func NewGoObject(val interface {}) Object {
  if reflect.ValueOf(val).Kind() == reflect.Func {
    // Functions have a special type
    return GoFunc{val: val}
  } else {
    return GoObject{val: val}
  }
}

/*
GoLX Type for an object implemented in Go
*/
func (obj GoObject) Type() Type {
  return GetGoType(reflect.TypeOf(obj.val))
}

/*
Retrieve methods on Go objects. Uses reflection to call 
*/
func (obj GoObject) Method(name string) Function {
  val := reflect.ValueOf(obj.val)
  method := val.MethodByName(name)
  if method.IsValid() {
    return GoFunc{val: method.Interface()}
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
Return Go methods as GoLX methods
*/
func (t GoType) Method(name string) Method {
  gomethod, exists := t.refType.MethodByName(name)
  if exists {
    return GoMethod(gomethod)
  } else
    return nil
  }
}

func (t GoType) Methods() []string {
  rt := t.refType
  mCount := rt.NumMethod()
  mList := make([]string, mCount, mCount)

  for i := range mList {
    method := rt.Method(i)
    mList[i] = method.Name()
  }

  return mList
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

func (f GoFunc) ObjectMethod(name string) Function {
  return GoObject(f).Method(name)
}

func (f GoFunc) FunctionType() FunctionType {
  ft, err := GetFuncType(reflect.TypeOf(f.val))
  if err != nil {
    panic("Only functions can be used as GoFunc objects")
  }
  return ft
}

func (f GoFunc) Type() Type {
  return Type(f.FunctionType())
}


func (f GoFunc) Call(args []Object) ([]Object, error) {
  fType := f.FunctionType()
  fVal := reflect.ValueOf(f.val)
  argVals := make([]reflect.Value, len(args))

  if len(args) != len(fType.Args()) {
    return []Object{}, errors.New("Incorrect number of arguments")
  }

  for i, arg := range args {
    if arg.Type() != fType.Args()[i] {
      return []Object{}, errors.New("Incorrect argument type for arg " + string(i))
    }

    argGoType, argIsGoType := fType.Args()[i].(GoType)

    if argIsGoType {
      argVals[i] = reflect.ValueOf(args[i].(GoObject).val).Convert(argGoType.refType)
    } else {
      argVals[i] = reflect.ValueOf(args[i])
    }
  }

  retVals := fVal.Call(argVals)

  rets := make([]Object, len(retVals))

  for i, retVal := range retVals {
    retObj, retIsObj := retVal.Interface().(Object)
    if retIsObj {
      rets[i] = retObj
    } else {
      rets[i] = NewGoObject(retVal.Interface())
    }
  }

  return rets, nil
}

func (m GoMethod) Apply(obj Object) (Function, error) {
  f := reflect.Method(m).Func

  inType := GetGoType(f.Type().In(0))

  if obj.Type() != inType {
    return nil, errors.New("Object is not the correct type for the method")
  }

  // It is much easier to let Go implement the actual closure
  return Obj.ObjectMethod(reflect.Method(m).Name())
}
