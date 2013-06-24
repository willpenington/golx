/*
Artnet Addressing

Translates to and from the 15 bit address encoding used in Art-Net packets to
the three 8 bit human readable numbers
*/
package artnet

import "fmt"

const (
  uniMask uint16 = 0x000F
  uniOffset uint = 0
  subMask uint16 = 0x00F0
  subOffset uint = 4
  netMask uint16 = 0x7F00
  netOffset uint = 8
)

type ArtnetAddress struct {
  universe uint8
  subnet uint8
  network uint8
}

func (addr ArtnetAddress) Universe() uint8 {
  return addr.universe
}

func (addr ArtnetAddress) Subnet() uint8 {
  return addr.subnet
}

func (addr ArtnetAddress) Network() uint8 {
  return addr.network
}

func (addr ArtnetAddress) String() string {
  return fmt.Sprintf("(%d,%d,%d)", addr.universe, addr.subnet, addr.network)
}

func NewArtnetAddress(universe uint8, subnet uint8, network uint8) ArtnetAddress {
  var addr ArtnetAddress

  addr.universe = universe
  addr.subnet = subnet
  addr.network = network

  return addr
}

func (addr ArtnetAddress) Encode() uint16 {
  uniPart := (uint16(addr.universe) << uniOffset) & uniMask
  subPart := (uint16(addr.subnet) << subOffset) & subMask
  netPart := (uint16(addr.network) << netOffset) & netMask
  return uniPart | subPart | netPart
}

func DecodeArtnetAddress(portAddr uint16) ArtnetAddress {
  var addr ArtnetAddress

  addr.universe = uint8((portAddr & uniMask) >> uniOffset)
  addr.subnet = uint8((portAddr & subMask) >> subOffset)
  addr.network = uint8((portAddr & netMask) >> netOffset)

  return addr
}
