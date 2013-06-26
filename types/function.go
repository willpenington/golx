package types

import (
	"errors"
	"reflect"
)

type FunctionType struct {
	Args []Type
	Rets []Type
}

type Function interface {
	Type() FunctionType
	Call(args []Object) ([]Object, error)
}

func (ft FunctionType) Name() string {

	name := "func("
	for i, t := range ft.Args {
		if i != 0 {
			name += ", "
		}
		name += t.Name()
	}
	name += ")"

	for i, t := range ft.Rets {
		if i != 0 {
			name += ","
		}
		name += " " + t.Name()
	}

	return name
}

func (ft FunctionType) Type() Type {
  return MetaType()
}

type functionImpl struct {
	function interface{}
	funcType FunctionType
}

func getType(valType reflect.Type) (Type, error) {
	// Build a new instance of the type
	sampleVal := reflect.New(valType).Elem().Interface()

	sample, isObj := sampleVal.(Object)

	if !isObj {
		return Type(nil), errors.New("Not a GoLX Object")
	}

	return sample.Type(), nil
}

func getArgTypes(funcType reflect.Type) ([]Type, error) {
	args := make([]Type, funcType.NumIn())

	for i := 0; i < funcType.NumIn(); i++ {
		arg, err := getType(funcType.In(i))
		args[i] = arg
		if err != nil {
			return []Type{}, errors.New("Arguments are not all Objects")
		}
	}

  return args, nil
}

func getRetTypes(funcType reflect.Type) ([]Type, error) {
	rets := make([]Type, funcType.NumOut())

	for i := 0; i < funcType.NumOut(); i++ {
		ret, err := getType(funcType.Out(i))
		rets[i] = ret
		if err != nil {
			return []Type{}, errors.New("Returns are not all values")
		}
	}

  return rets, nil
}

func NewFunction(gofunc interface{}) (Function, error) {
	funcVal := reflect.ValueOf(gofunc)
	funcType := funcVal.Type()

	if funcVal.Kind() != reflect.Func {
		return Function(nil), errors.New("Not a function")
	}

	args, err := getArgTypes(funcType)

  if err != nil {
    return Function(nil), err
  }

	rets, err := getRetTypes(funcType)

  if err != nil {
    return Function(nil), err
  }

	ft := FunctionType{Args: args, Rets: rets}
	fImpl := functionImpl{funcType: ft, function: gofunc}

  return fImpl, nil
}

func (fImpl functionImpl) Type() FunctionType {
	return fImpl.funcType
}

func (fImpl functionImpl) Call(args []Object) ([]Object, error) {
	fVal := reflect.ValueOf(fImpl)
	argVals := make([]reflect.Value, len(args))

	for i, arg := range args {
		if arg.Type() != fImpl.Type().Args[i] {
			return []Object{}, errors.New("Type Error: Argument does not match" +
				"function type")
		}
		argVals[i] = reflect.ValueOf(args[i])
	}

	retVals := fVal.Call(argVals)

	rets := make([]Object, len(retVals))

	for i, retVal := range retVals {
		rets[i] = retVal.Interface().(Object)
	}

	return rets, nil
}
