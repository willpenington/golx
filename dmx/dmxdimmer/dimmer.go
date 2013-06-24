/*
Basic single channel dimmer fixture
*/
package dmxdimmer

import (
  "golx/fixture"
  "golx/dmx"
  "golx/dmx/dmxintensity"
  "golx/data/intensity"
)

type DMXDimmer struct {
  attr *dmxintensity.DMXIntensity
}

func NewDMXDimmer() *DMXDimmer {
  dimmer := new(DMXDimmer)
  dimmer.attr = dmxintensity.NewDMXIntensity(dimmer)
  return dimmer
}

func (dimmer *DMXDimmer) Attributes() map[string] fixture.Attribute {
  return map[string] fixture.Attribute{"intensity": dimmer.attr}
}

func (dimmer *DMXDimmer) Intensity() *dmxintensity.DMXIntensity {
  return dimmer.attr
}

func (dimmer *DMXDimmer) Output() chan dmx.DMXValue {
  return dimmer.attr.DMXOut().Output()
}

func (dimmer *DMXDimmer) Value() intensity.Intensity {
  return dimmer.Value()
}

func (dimmer *DMXDimmer) SetValue(val intensity.Intensity) {
  dimmer.attr.SetValue(val)
}
