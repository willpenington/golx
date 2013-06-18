/*
Provides a Go Channel that can send or recieve values for a single DMX Channel
in a universes. Needs rewriting.
*/
package dmx

import "fmt"

type DMXChannel struct {
  universe *DMXUniverse
  channelNumber int
  replace chan bool
}

func newDMXChannel(universe *DMXUniverse, channelNumber int) *DMXChannel {
  channel := new(DMXChannel)
  channel.universe = universe
  channel.channelNumber = channelNumber

  channel.replace = make(chan bool)
  go func() { _ = <-channel.replace }()

  return channel
}

func (channel *DMXChannel) setValue(val DMXValue) {
  channel.universe.setValue(channel.channelNumber, val)
}

func (channel *DMXChannel) GetValue() DMXValue {
  return channel.universe.getValue(channel.channelNumber)
}

func (channel *DMXChannel) Patch(input chan DMXValue) {

  fmt.Println("Patching channel ", channel)
  fmt.Println("Patching with channel.replace == nil as ", channel.replace == nil)
  channel.replace <- true // Make sure any other goroutine has stopped before patching

  go func() {
    for {
      select {
      case _ = <-channel.replace:
        return
      case val := <-input:
        channel.setValue(val)
      }
    }
  }()
}

func (channel *DMXChannel) Unpatch() {
  channel.replace <- true
  go func() { _ = <-channel.replace }()
}
