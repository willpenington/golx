/*
Fixture

A fixture represents a logical grouping of Attributes and parameters.
*/
package fixture

import "golx/attribute"

type Fixture interface {
  Attribute(name string) attribute.Attribute
  Attributes() []attribute.Attribute
  AttributeLabes() []string
}


