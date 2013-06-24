/*
DMX Value types
*/
package dmx

import "fmt"

// A Value recieved from or destined for a DMX input or output
type DMXValue uint8
// A DMX Universe as formated in a NULL START DMX Packet
type DMXFrame []DMXValue

func (val DMXValue) String() string {
  return fmt.Sprintf("<DMX Value %d>", uint8(val))
}
