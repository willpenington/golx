/*
Represents a DMX output from an Attribute

Provides common functionallity 
*/
package dmx

import (
  "time"
  "golx/fixture"
  "golx/dmx"
)

const (
  defaultValue dmx.DMXValue = 0
  timeoutLimit = 1 * time.Second
)

type DMXParam struct {
  value dmx.DMXValue
  attr *fixture.Attribute
  output chan dmx.DMXValue
}

func NewDMXParam(attr *fixture.Attribute) *DMXParam {
  param := new(DMXParam)

  param.value = defaultValue
  param.attr = attr

  param.output = make(chan dmx.DMXValue)

  return param
}

func (param *DMXParam) Attribute() *fixture.Attribute {
  return param.attr
}

func (param *DMXParam) Value() dmx.DMXValue {
  return param.value
}

func (param *DMXParam) SetValue(val dmx.DMXValue) {
  param.value = val

  select {
  case param.output <- val:
  default:
  }
}

func (param *DMXParam) Output() chan dmx.DMXValue {
  return param.output
}
