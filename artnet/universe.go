package artnet

import (
  "golx/dmx"
  "time"
  "net"
  "fmt"
)

type ArtnetUniverse struct {
  address ArtnetAddress
  input chan dmx.DMXFrame
  output chan dmx.DMXFrame
  netInput chan *artDmx
  sendHold chan bool

  sequence chan uint8
  quitSequence chan bool

  physicalSend uint8
  physicalRecv uint8

  rateLimit time.Duration
  keepAlive time.Duration
}

var universes map[uint16] *ArtnetUniverse
var universesLock chan bool

func init() {
  universes = make(map[uint16] *ArtnetUniverse)
  universesLock = make(chan bool, 1)
}

func GetArtnetUniverse(address ArtnetAddress) *ArtnetUniverse {
  universesLock <- true

  val, ok := universes[address.Encode()]

  if !ok {
    val = newArtnetUniverse(address)
    universes[address.Encode()] = val
  }

  _ = <-universesLock

  return val
}

func newArtnetUniverse(address ArtnetAddress) *ArtnetUniverse {
  universe := new(ArtnetUniverse)

  universe.address = address
  universe.input = make(chan dmx.DMXFrame)
  universe.output = make(chan dmx.DMXFrame)
  universe.netInput = make(chan *artDmx)
  universe.sendHold = make(chan bool)

  universe.quitSequence = make(chan bool)
  universe.sequence = sequence(universe.quitSequence)

  universe.physicalSend = 0
  universe.physicalRecv = 0

  universe.rateLimit = 25 * time.Millisecond
  universe.keepAlive = 4 * time.Minute

  go universe.netListen()
  go universe.netSend()

  return universe
}

func (u *ArtnetUniverse) String() string {
  return "[Artnet Universe @ " + u.address.String() + "]"
}

func (u *ArtnetUniverse) Input() chan dmx.DMXFrame {
  return u.input
}

func (u *ArtnetUniverse) Output() chan dmx.DMXFrame {
  return u.output
}

func (u *ArtnetUniverse) Address() ArtnetAddress {
  return u.address
}

func (u *ArtnetUniverse) LocalPhysical() uint8 {
  return u.physicalSend
}

func (u *ArtnetUniverse) SetLocalPhysical(physical uint8) {
  u.physicalSend = physical
}

func (u *ArtnetUniverse) RemotePhysical() uint8 {
  return u.physicalRecv
}

func (u *ArtnetUniverse) netListen() {
  for packet := range u.netInput {
    u.physicalRecv = packet.physical
    u.output <- packet.frame
  }
}

func (universe *ArtnetUniverse) sendFrame(data dmx.DMXFrame) {
  broadcastAddr, err := net.ResolveUDPAddr("udp", "2.255.255.255:6465")

  if err != nil {
    return
  }

  artdmx := new(artDmx)
  artdmx.address = universe.address
  artdmx.physical = universe.physicalSend
  artdmx.sequence = <-universe.sequence
  artdmx.frame = data


  sendArtDmx(artdmx, broadcastAddr)
}

func sequence(quit chan bool) chan uint8 {
  seq := uint8(1)
  c := make(chan uint8)

  go func() {
    for {
      select {
      case c <- seq:
        seq = (seq % 255) + 1
      case _ = <-quit:
        return
      }
    }
  }()

  return c
}

func (u *ArtnetUniverse) rateLimitInput() chan dmx.DMXFrame {
  tick := time.After(0)
  out := make(chan dmx.DMXFrame)
  last := dmx.DMXFrame(nil)
  newData := false

  go func() {
    for {
      select {
      case _ = <-tick:
        if newData {
          out <- last
          newData = false
        } else {
          last = <-u.input
          fmt.Println("Artnet got data")
          out <- last
        }
        tick = time.After(u.rateLimit)
      case last = <-u.input:
        newData = true
        fmt.Println("Artnet got data")
      }
    }
  }()

  return out
}

func (u *ArtnetUniverse) netSend() {
  input := u.rateLimitInput()
  data := dmx.DMXFrame(nil)
  var keepAlive <-chan time.Time

  for {
    select {
    case data = <-input:
      u.sendFrame(data)
      keepAlive = time.After(u.keepAlive)
    case _ = <-keepAlive:
      input <- data
    case hold := <-u.sendHold:
      if hold {
        keepAlive = nil
      }
    }
  }
}

func (u *ArtnetUniverse) StopSending() {
  u.sendHold <- true
}
