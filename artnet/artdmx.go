/*
ArtDmx Packet support

Functions for handling artDmx Packets and representing them as a stream of
byte slices
*/
package artnet

import (
  "errors"
  "bytes"
  "net"
  "fmt"
  "encoding/binary"
  "io"
  "time"
  "golx/dmx"
)

/*
Data (excluding header) included in the ArtDmx Packet
*/
type artDmx struct {
  sequence uint8
  physical uint8
  net uint8
  sub uint8
  universe uint8
  frame dmx.DMXFrame
}

// Outputs to send DMX data from recieved packets to
var inputUniverseChannels map[uint16] chan dmx.DMXFrame = nil

func handleArtDmx(r *bytes.Buffer, source *net.UDPAddr) error {

  artdmx := parseArtDmx(r)

  // Use the ArtNet encoding to map the ArtNet address to an integer
  portAddr := BuildPortAddr(artdmx.net, artdmx.sub, artdmx.universe)
  recipient, ok := inputUniverseChannels[portAddr]

  // Send the frame to the recipient channel if exists
  if ok && recipient != nil {
    recipient <- artdmx.frame
  }

  return nil
}

// Parse a byte stream to a artDmx packet struct
func parseartDmx(r *io.Reader) *artDmx {
  artdmx := new(artDmx)

  var portAddr uint16
  var length uint16

  binary.Read(r, binary.LittleEndian, &(artdmx.sequence))
  binary.Read(r, binary.LittleEndian, &(artdmx.physical))
  binary.Read(r, binary.LittleEndian, &portAddr)
  binary.Read(r, binary.BigEndian, &length)

  artdmx.net = GetNet(portAddr)
  artdmx.sub = GetSubnet(portAddr)
  artdmx.universe = GetUniverse(portAddr)

  artdmx.frame = make(dmx.DMXFrame, length)

  // Copy the bytes into the DMX frame one by one
  for i := uint16(0); i < length; i++ {
    level, _ := r.ReadByte()
    artdmx.frame[i] = dmx.DMXValue(level)
  }

  return artdmx
}

// Build and send a new ArtDmx packet
func sendArtDmx(artdmx *artDmx, addr *net.UDPAddr) error {

  buf := bytes.NewBuffer(make([]byte, 0))

  // Write standard header values
  WriteHeader(buf, OpDmx)

  // Encode the port address
  portAddr := BuildPortAddr(artdmx.net, artdmx.sub, artdmx.universe)

  // Write the binary values onto the buffer
  binary.Write(buf, binary.LittleEndian, artdmx.sequence)
  binary.Write(buf, binary.LittleEndian, artdmx.physical)
  binary.Write(buf, binary.LittleEndian, portAddr)
  binary.Write(buf, binary.BigEndian, uint16(len(artdmx.frame)))

  // Copy the DMX values into the buffer
  for i := 0; i < len(artdmx.frame); i++ {
    buf.WriteByte(byte(artdmx.frame[i]))
  }

  // Use the central network connection to dispatch the packet
  SendPacket(buf.Bytes(), addr)

  return nil
}

// Get a channel that returns all DMX data sent to the Art-Net address
func InputArtnetUniverse(network uint8, subnet uint8, universe uint8) (chan dmx.DMXFrame, error) {

  // Initialise the input channel list if needed 
  if inputUniverseChannels == nil {
    inputUniverseChannels = make(map[uint16] chan dmx.DMXFrame)
  }

  // Build the Art-Net encoded address for lookup
  portAddr := BuildPortAddr(network, subnet, universe)

  // Check if an output channel already exists
  _, exists := inputUniverseChannels[portAddr]

  if exists {
    return nil, errors.New("Universe already has a channel")
  }

  // Build store and return a new channel
  dmx := make(chan dmx.DMXFrame)
  inputUniverseChannels[portAddr] = dmx
  return dmx, nil
}

/*
Build a channel that will send DMX Data it recieves to the specified 
Art-Net address 
*/
func OutputArtnetUniverse(network uint8, subnet uint8, universe uint8, physical uint8) (chan dmx.DMXFrame, error) {

  // For the moment, all packets are sent to broadcast. When Art-Pol and
  // subscriptions are properly implemented this will be done with Unicast to
  // the correct address.
  broadcastAddr, err := net.ResolveUDPAddr("udp", "2.255.255.255:6465")
  if err != nil {
    return nil, err
  }

  c := make(chan dmx.DMXFrame)

  // Keepalive
  slowTick := make(chan time.Time)
  // Rate limiting
  fastTick := make(chan time.Time)

  sequence := uint8(0)

  // Convenient closure for sending each frame automatically
  sendFrame := func(frame dmx.DMXFrame) {
    // Update the sequence byte, rolling over from 255 to 1
    sequence = (sequence % 255) + 1

    // Build the packet
    artdmx := new(artDmx)
    artdmx.universe = universe
    artdmx.sub = subnet
    artdmx.net = network
    artdmx.sequence = sequence
    artdmx.physical = physical
    artdmx.frame = frame

    SendartDmx(artdmx, broadcastAddr)
  }

  go func() {
    frame := dmx.DMXFrame(nil)
    // Limit rate to at most 40 frames a second
    for _ = <-fastTick {
      select {
      case frame := <-c:
        sendFrame(frame)
        // Set timeouts
        fastTick = time.After(25 * time.Millisecond)
        slowTick = time.After(4 * time.Minutes)
      case _ = <-slowTick:
        // slowTick only gets data after frame has recieved data and therefore
        // frame should not be nil at this point
        c <- frame
      }
    }
  }()

  // Provide the first fast tick to kick start the listner
  fastTick <- time.Now()

  return c, nil
}
