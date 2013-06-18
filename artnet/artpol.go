/*
Handle and Send ArtPol packets

Currently a placeholder
*/
package artnet

import (
  "net"
  "fmt"
  "bytes"
)

type ArtPol struct {
  TalkToMe uint8;
  Priority uint8;
}

func NewArtPol() *ArtPol {
  artpol := new(ArtPol)
  artpol.TalkToMe = 0
  artpol.Priority = 0
  return artpol
}

func HandleArtPol(r *bytes.Buffer, source *net.UDPAddr) error {
  fmt.Println("Got an ArtPol packet")
  return nil
}
