/*
Provides a Go Channel that can send or recieve values for a single DMX Channel
in a universes. Needs rewriting.
*/
package dmx

import "fmt"

type DMXChannel struct {
  universe *DMXUniverse
  channelNumber int
  output chan DMXValue
  input chan DMXValue
}

func newDMXChannel(universe *DMXUniverse, channelNumber int) *DMXChannel {
  channel := new(DMXChannel)
  channel.universe = universe
  channel.channelNumber = channelNumber
  channel.input = nil
  channel.output = make(chan DMXValue)

  return channel
}

func (channel *DMXChannel) String() string {
  return fmt.Sprintf("[%s.%d]", channel.universe.String(), channel.channelNumber)
}

func (channel *DMXChannel) Input() chan DMXValue {
  if channel.input == nil {
    channel.buildInput()
  }

  return channel.input
}

func (channel *DMXChannel) Output() chan DMXValue {
  return channel.output
}

func (channel *DMXChannel) setValue(val DMXValue) {
  channel.universe.setValue(channel.channelNumber, val)
}

func (channel *DMXChannel) Value() DMXValue {
  return channel.universe.getValue(channel.channelNumber)
}

func (channel *DMXChannel) buildInput() {

  channel.input = make(chan DMXValue)

  go func() {
    for val := range channel.input {
      fmt.Println("DMXChannel got data: ", val)
      channel.setValue(val)
    }
  }()
}

func (channel *DMXChannel) sendOutput() {
  channel.output <- channel.Value()
}
