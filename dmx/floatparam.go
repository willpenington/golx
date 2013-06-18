/*
Floating Point DMX Parameter

Maps the 0 - 1 range of a floating point value onto the range of a DMX value
(0 - 255). Needs rewriting to use channels
*/
package dmx

import "math"

type DMXFloatParam struct {
  dmx DMXParam
}

func NewDMXFloatParam(dmx DMXParam) *DMXFloatParam {
  param := new(DMXFloatParam)
  param.dmx = dmx
  return param
}

func (param *DMXFloatParam) SetValue(val float64) {
  gated := math.Min(math.Max(0, val), 1)

  dmxVal := DMXValue(gated * 255)
  param.dmx.SetValue(dmxVal)
}

func (param *DMXFloatParam) Value() float64 {
  return float64(param.dmx.Value()) / 255
}
