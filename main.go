/*
GOLX

This file currently sets up and tears down the implemented infrastructure and
tries to send some test data. At the moment it is primarily for debugging.
*/
package main

import (
  "fmt"
  "time"
	"golx/artnet"
	"golx/dmx"
  "golx/patch"
  "golx/dmx/dmxdimmer"
  // "golx/dmx/dmxfixture"
  "golx/data/intensity"
)

type ticker bool

func (t ticker) Output() chan dmx.DMXValue {
  c := make(chan dmx.DMXValue)
  tick := time.Tick(1 * time.Second)

  go func() {
    for now := range tick {
      fmt.Println("Ticking!")
      c <- dmx.DMXValue(now.Second())
    }
  }()

  return c
}

func main() {
	anUniv := artnet.GetArtnetUniverse(artnet.NewArtnetAddress(0,0,0))
	universe := dmx.NewDMXUniverse()

  patch.Patch(universe, anUniv)

  dimmers := make([]*dmxdimmer.DMXDimmer, 2)

  for i := 1; i < 3; i++ {
    dimmer := dmxdimmer.NewDMXDimmer()
    err := patch.Patch(dimmer.Intensity().DMXOut(), universe.GetChannel(i))

    if err != nil {
      fmt.Println(err.Error())
    }

    dimmers[i - 1] = dimmer
  }

  // t := ticker(true)

  // err := patch.Patch(t, universe.GetChannel(7))

  // if err != nil {
  //  fmt.Println(err.Error())
  //}

  // universe.GetChannel(1).Input() <- dmx.DMXValue(3)

  fmt.Println("Getting a new channel")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c1 := dimmers[0].Intensity().Input()
  _ = <-time.After(1 * time.Second)

  fmt.Println("Setting intensity on first channel to 0.5")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c1 <- intensity.Intensity(0.5)
  _ = <-time.After(1 * time.Second)

  fmt.Println("Setting intensity on first channel to 0.6")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c1 <- intensity.Intensity(0.6)
  _ = <-time.After(1 * time.Second)

  fmt.Println("Getting another new channel")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c2 := dimmers[0].Intensity().Input()
  _ = <-time.After(1 * time.Second)

  fmt.Println("Setting intensity on second channel to 0.7")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c2 <- intensity.Intensity(0.7)
  _ = <-time.After(1 * time.Second)

  fmt.Println("Setting intensity on first channel to 0.3")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c1 <- intensity.Intensity(0.3)
  _ = <-time.After(1 * time.Second)

  fmt.Println("Releasing second channel")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  close(c2)
  _ = <-time.After(1 * time.Second)

  fmt.Println("Setting intensity on first channel to 0.9")
  fmt.Println("Value for intensity: ", dimmers[0].Intensity().Value())
  c1 <- intensity.Intensity(0.9)
  _ = <-time.After(1 * time.Second)

  fmt.Println("Final value for intensity: ", dimmers[0].Intensity().Value())

  // param := dmxfixture.NewDMXParam(nil)
  // patch.Patch(param, universe.GetChannel(10))
  // param.SetValue(dmx.DMXValue(100))
  // universe.GetChannel(11).Input() <- dmx.DMXValue(101)

  _ = <-time.After(1 * time.Second)
}
