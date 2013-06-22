/*
GOLX

This file currently sets up and tears down the implemented infrastructure and
tries to send some test data. At the moment it is primarily for debugging.
*/
package main

import (
	"fmt"
	"golx/artnet"
	"golx/dmx"
  "golx/patch"
	"time"
)

func main() {
	anUniv := artnet.GetArtnetUniverse(artnet.NewArtnetAddress(0,0,0))
	universe := dmx.NewDMXUniverse()

  patch.Patch(universe, anUniv)

	tick := time.Tick(1 * time.Second)
	channelNumber := 0

	for now := range tick {
		channelNumber += 1

    channel := universe.GetChannel(channelNumber)

    channel.Input() <- dmx.DMXValue(now.Second())

		fmt.Println("Value sent")

		if channelNumber == 512 {
			return
		}
	}

}
