package types

type FunctionType struct {
	args []Type
	rets []Type
}

func NewFunctionType(args, rets []Type) FunctionType {
  return FunctionType{args: args, rets: rets}
}

func (ft FunctionType) Args() []Type {
  return ft.args
}

func (ft FunctionType) Rets() []Type {
  return ft.rets
}

type Function interface {
	Type() Type
  FunctionType() FunctionType
	Call(args []Object) ([]Object, error)
}

func (ft FunctionType) Name() string {

	name := "func("
	for i, t := range ft.Args() {
		if i != 0 {
			name += ", "
		}
		name += t.Name()
	}
	name += ")"

	for i, t := range ft.Rets() {
		if i != 0 {
			name += ","
		}
		name += " " + t.Name()
	}

	return name
}


