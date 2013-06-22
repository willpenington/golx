/*
OpCode constants and lookups

This file defines the OpCode type, constants for every OpCode in the Art-Net specification
and the lookup function for OpCode Handlers
*/

package artnet

import (
  "errors"
  "net"
  "bytes"
)

type Opcode uint16

const (
	OpPoll              Opcode = 0x2000
	OpPollReply         Opcode = 0x2100
	OpDiagData          Opcode = 0x2300
	OpCommand           Opcode = 0x2400
	OpOutput            Opcode = 0x5000
  OpDmx               Opcode = 0x5000
	OpNzs               Opcode = 0x5100
	OpAddress           Opcode = 0x6000
	OpInput             Opcode = 0x7000
	OpTodRequest        Opcode = 0x8000
	OpTodData           Opcode = 0x8100
	OpTodControl        Opcode = 0x8200
	OpRdm               Opcode = 0x8300
	OpRdmSub            Opcode = 0x8400
	OpVideoSetup        Opcode = 0xa010
	OpVideoPalette      Opcode = 0xa020
	OpVideoData         Opcode = 0xa040
	OpMacMaster         Opcode = 0xf00
	OpMacSlave          Opcode = 0xf100
	OpFirmwareMaster    Opcode = 0xf200
	OpFirmwareReply     Opcode = 0xf300
	OpFileTnMaster      Opcode = 0xf400
	OpFileFnMaster      Opcode = 0xf500
	OpFileFnReply       Opcode = 0xf600
	OpIpProg            Opcode = 0xf800
	OpIpProgReply       Opcode = 0xf900
	OpMedia             Opcode = 0x9000
	OpMediaPatch        Opcode = 0x9100
	OpMediaControl      Opcode = 0x9200
	OpMediaControlReply Opcode = 0x9300
	OpTimeCode          Opcode = 0x9700
	OpTimeSync          Opcode = 0x9800
	OpTrigger           Opcode = 0x9900
	OpDirectory         Opcode = 0x9a00
	OpDirectoryReply    Opcode = 0x9b00
)

func HandlerByOpcode(opcode Opcode) (func(r *bytes.Buffer, source *net.UDPAddr) error,  error) {
  switch opcode {
  case OpPoll:
    return HandleArtPol, nil
  case OpDmx:
    return handleArtDmx, nil
  }
  return nil, errors.New("Opcode not implemented")
}
