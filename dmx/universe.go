/*
Abstracts a DMX Universe across a physical input and output and stores the
most recent values
*/
package dmx

import "math"

const (
  UniverseSize int = 512
)

type DMXUniverse struct {
  data DMXFrame
  output chan DMXFrame
  input chan DMXFrame
  channels [](*DMXChannel)
}

func NewDMXUniverse() *DMXUniverse {
  universe := new(DMXUniverse)
  universe.output = make(chan DMXFrame)
  universe.input = make(chan DMXFrame)
  universe.data = make(DMXFrame, UniverseSize)
  universe.channels = make([](*DMXChannel), UniverseSize)
  universe.buildChannels()
  return universe
}

func (u *DMXUniverse) listen() {
  for frame := range u.input {
    for i := 1; i <= math.Min(len(u.data), len(frame)); i++ {
      setValue(i, frame[i])
    }
  }
}

func (u *DMXUniverse) buildChannels() {
  for i := 0; i < UniverseSize; i++ {
    u.channels[i] = newDMXChannel(u, i + 1)
  }
}

func (u *DMXUniverse) GetChannel(channelNumber int) *DMXChannel {
  return u.channels[channelNumber]
}

func (u *DMXUniverse) Channels() *[]*DMXChannel {
  return &(u.channels)
}

func (u *DMXUniverse) setValue(channel int, value DMXValue) {
  if value != u.getValue(channel) {
    u.data[channel - 1] = value
    u.output <- u.data
    u.GetChannel(channel).sendValue()
  }
}

func (u *DMXUniverse) getValue(channel int) DMXValue {
  return u.data[channel - 1]
}

func (u *DMXUniverse) InputChannel() chan DMXFrame {
  return u.input
}

func (u *DMXUniverse) OutputChannel() chan DMXFrame {
  return u.output
}
