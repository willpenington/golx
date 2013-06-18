/*
Art-Net common functions and connection management
*/
package artnet

import (
  "net"
  "bytes"
  "encoding/binary"
  "fmt"
  "errors"
)

// Data common to all ArtNet packets
type ArtNetHeader struct {
  ID [8]byte;
  OpCode Opcode;
  ProtVerHi uint8;
  ProtVerLo uint8;
}

// Write the standard packet header onto a byte stream
func WriteHeader(buf *bytes.Buffer, opcode Opcode) {
  buf.WriteString("Art-Net\x00")
  binary.Write(buf, binary.LittleEndian, uint16(opcode))
  binary.Write(buf, binary.LittleEndian, uint8(0))
  binary.Write(buf, binary.LittleEndian, uint8(14))
}

// Handle incoming packets
func HandlePacket(buf *bytes.Buffer, source *net.UDPAddr) {
  id, err := buf.ReadBytes(0)

  // Parse header data

  if err != nil {
    err = errors.New("Error in parsing id")
  }

  var opCode uint16
  var protVerHi uint8
  var protVerLo uint8

  err = binary.Read(buf, binary.LittleEndian, &opCode)

  if err != nil {
    fmt.Println("Error in parsing opCode: ", err)
    err = nil
  }

  err = binary.Read(buf, binary.LittleEndian, &protVerHi)

  if err != nil {
    fmt.Println("Error in parsing protVerHi: ", err)
    err = nil
  }

  err = binary.Read(buf, binary.LittleEndian, &protVerLo)

  if err != nil {
    fmt.Println("Error in parsing protVerLo: ", err)
    err = nil
  }


  if err != nil {
    fmt.Println("Error in parsing header: ", err)
  }

  // Check standard header data meets the specification
  if string(id) == "Art-Net\x00" && protVerHi == 0 && protVerLo >= 14 {
    // Look up handler function
    handler, err := HandlerByOpcode(Opcode(opCode))

    if err == nil {
      // Use OpCode specific handler to process the packet
      handler(buf, source)
    } else  {
      fmt.Println("Error: ", err)
    }
  } else {
    fmt.Println("Packet Header Invalid")
    fmt.Printf("ID: %s  Valid %t\n", string(id), string(id) == "Art-Net\x00")
    fmt.Printf("Opcode: %X\n", opCode)
    fmt.Printf("ProtVerHi: %d Valid %t\n", protVerHi, protVerHi == 0)
    fmt.Printf("ProtVerLo: %d Valid %t\n", protVerLo, protVerLo >= 14)
  }

}
