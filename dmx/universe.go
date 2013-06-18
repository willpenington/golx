/*
Abstracts a DMX Universe across a physical input and output and stores the
most recent values
*/
package dmx

const (
  UniverseSize int = 512
)

type DMXUniverse struct {
  data DMXFrame
  output chan DMXFrame
  channels [](*DMXChannel)
}

func NewUniverse(output chan DMXFrame) *DMXUniverse {
  universe := new(DMXUniverse)
  universe.output = output
  universe.data = make(DMXFrame, UniverseSize)
  universe.channels = make([](*DMXChannel), UniverseSize)
  universe.buildChannels()
  return universe
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
  u.data[channel - 1] = value
  u.output <- u.data
}

func (u *DMXUniverse) getValue(channel int) DMXValue {
  return u.data[channel - 1]
}
