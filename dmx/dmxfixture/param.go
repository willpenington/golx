/*
Represents a DMX output from an Attribute

Provides common functionallity for attributes that provide their output as DMX
values by converting method calls to a channel output and caching the current
value for display.
*/
package dmxfixture

import (
  "fmt"
  "golx/fixture"
  "golx/dmx"
  "golx/patch/chanutil"
)

const (
  defaultValue dmx.DMXValue = 0
)

type DMXParam struct {
  value dmx.DMXValue
  attr fixture.Attribute
  output chan dmx.DMXValue
  publicOutput chan dmx.DMXValue
}

func NewDMXParam(attr fixture.Attribute) *DMXParam {
  param := new(DMXParam)

  param.value = defaultValue
  param.attr = attr

  /*
  Use DeliverWhenPossible so that SetValue is not blocking
  */
  param.publicOutput = make(chan dmx.DMXValue)
  param.output = make(chan dmx.DMXValue)
  chanutil.DeliverWhenPossible(param.output, param.publicOutput)

  return param
}

/*
The Attribute that created and updates this parameter
*/
func (param *DMXParam) Attribute() fixture.Attribute {
  return param.attr
}

/*
The last value the paramaeter was set to. Intended to be used for display only.
Patch to the output for all other uses.
*/
func (param *DMXParam) Value() dmx.DMXValue {
  return param.value
}

/*
Set the parameter to a new value. This should only be set by the attribute that
created the parameter.
*/
func (param *DMXParam) SetValue(val dmx.DMXValue) {
  fmt.Println("DMXParam got data")
  param.value = val

  param.output <- val
}

/*
Get the channel that this parameter outputs on
*/
func (param *DMXParam) Output() chan dmx.DMXValue {
  return param.publicOutput
}
