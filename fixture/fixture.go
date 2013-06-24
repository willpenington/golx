package fixture

type Fixture interface {
  Attributes() map[string] Attribute
}
