/*
Port Address helpers

Translates to and from the 15 bit address encoding used in Art-Net packets to
the three 8 bit human readable numbers
*/
package artnet

const (
  uniMask uint16 = 0x000F
  uniOffset uint = 0
  subMask uint16 = 0x00F0
  subOffset uint = 4
  netMask uint16 = 0x7F00
  netOffset uint = 8
)

func BuildPortAddr(net uint8, sub uint8, uni uint8) uint16 {
  uniPart := (uint16(uni) << uniOffset) & uniMask
  subPart := (uint16(sub) << subOffset) & subMask
  netPart := (uint16(net) << netOffset) & netMask
  return uniPart | subPart | netPart
}

func GetUniverse(addr uint16) uint8 {
  return uint8((addr & uniMask) >> uniOffset)
}

func GetSubnet(addr uint16) uint8 {
  return uint8((addr & subMask) >> subOffset)
}

func GetNet(addr uint16) uint8 {
  return uint8((addr & netMask) >> netOffset)
}
