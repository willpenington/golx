/*
GOLX

This file currently sets up and tears down the implemented infrastructure and
tries to send some test data. At the moment it is primarily for debugging.
*/
package main

import (
  "golx/artnet"
  "golx/dmx"
  "time"
  "fmt"
)

func main() {
  artnet.Start()

  dmxchan, _ := artnet.OutputArtnetUniverse(0, 0, 0, 0)
  universe := dmx.NewUniverse(dmxchan)

  tick := time.Tick(1 * time.Second)
  channelNumber := 0

  for now := range tick {
    channelNumber += 1
    fmt.Println("Patching channel ", channelNumber)
    channel := universe.GetChannel(channelNumber)
    fmt.Println("Channel retrieved")
    attr := dmx.NewDMXAttribute("blah")
    fmt.Println("Attribute created")
    c := attr.Subscribe()
    fmt.Println("Subscription created")
    channel.Patch(c)
    fmt.Println("Patched")
    attr.SetValue(dmx.DMXValue(now.Second()))
    fmt.Println("Value sent")

    if channelNumber == 512 {
      return
    }
  }

}
