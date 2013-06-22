package fixture

type ParamValue interface {}

type Parameter interface {
  Value() ParamValue
  SetValue(ParamValue)
  Attribute() *Attribute
  Output chan ParamValue
}
