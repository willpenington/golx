package dmxintensity

import (
  "fmt"
  "golx/fixture"
  "golx/fixture/mixer"
  "golx/dmx"
  "golx/dmx/dmxfixture"
  "golx/data/intensity"
)

type DMXIntensity struct {
  fixture fixture.Fixture
  param *dmxfixture.DMXParam
  mixer *mixer.LTPMixer

  input chan intensity.Intensity
  value intensity.Intensity

  stop chan bool
}



func NewDMXIntensity(fixture fixture.Fixture) *DMXIntensity {
  attr := new(DMXIntensity)
  attr.fixture = fixture
  attr.param = dmxfixture.NewDMXParam(attr)

  attr.input = make(chan intensity.Intensity)
  attr.mixer, _ = mixer.NewLTPMixer(attr.input, intensity.Intensity(0))
  attr.value = 0

  go func() {
    for {
      fmt.Println("Waiting for input in intensity")
      select {
      case val := <-attr.input:
        fmt.Println("Got input in intensity")
        attr.value = val
        go attr.param.SetValue(dmx.DMXValue(float64(val) * float64(255)))
        fmt.Println("Done blocking in intensity")
      case _ = <-attr.stop:
        return
      }
    }
  }()

  return attr
}

func (attr *DMXIntensity) Fixture() fixture.Fixture {
  return attr.fixture
}

func (attr *DMXIntensity) SetValue(val intensity.Intensity) {
  attr.input <- val
}

func (attr *DMXIntensity) Value() intensity.Intensity {
  return attr.value
}

func (attr *DMXIntensity) Input() chan intensity.Intensity {
  fmt.Println("Adding new channel")
  c := make(chan intensity.Intensity)
  fmt.Println("Channel built")
  attr.mixer.AddInput(c)
  fmt.Println("Channel added to mixer")
  fmt.Println("Returning channel")
  return c
}

func (attr *DMXIntensity) Parameters() map[string] fixture.Parameter {
  return map[string] fixture.Parameter{"intensity": attr.param}
}

func (attr *DMXIntensity) DMXOut() *dmxfixture.DMXParam {
  return attr.param
}
