/*
Attributes

Interfaces for dealing with attributes of lighting fixtures
*/
package attribute

import "golx/fixture"

/* 
Values that can be used to set an interface. Currently, this can be any go
type but it improves readability and makes future development easier if it
is a type
*/
type AttrValue interface{}

/*
Attributes store their current value and provide a channel that can be used to
set new values.

Attributes are defined as an interface so that functions that deal with them
can set the value without being aware that they may be behind a mixer.
*/
type Attribute interface {
  Fixture() fixture.Fixture
  Value() AttrValue
  GetInput() chan AttrValue
}
