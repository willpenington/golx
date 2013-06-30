package types

import (
  "reflect"
  "errors"
)

/*
Wraps a Go channel of GoLX objects of the same type
*/
type Port struct {
	portType PortType

	setPatchChan chan interface {}
	stop chan bool
}

/*
Type object for Port
*/
type PortType struct {
	valType Type
  broadcast bool
  input bool
  output bool
}

/*
Creates a port that will read/write data off the channel c
*/
func NewPort(c interface {}, t Type, broadcast bool) (Port, error) {
	stopChan := make(chan bool)
  setPatchChan := make(chan interface {})

  internal := reflect.ValueOf(c)
  stop := reflect.ValueOf(stopChan)
  setPatch := reflect.ValueOf(setPatchChan)
  patch := reflect.ValueOf(nil)

  if internal.Kind() != reflect.Chan {
    return Port{}, errors.New("c is not a chan")
  }

  // Listing aliases
  var internalRecvDefault reflect.Value
  if internal.Type().ChanDir() != reflect.SendDir {
    internalRecvDefault = internal
  } else {
    internalRecvDefault = reflect.ValueOf(nil)
  }
  patchRecvDefault := reflect.ValueOf(nil)

  internalRecv := internalRecvDefault
  patchRecv := patchRecvDefault

  // Dummy channel
  neverListen := reflect.ValueOf(make(chan interface {}))

  // Sending aliases
  internalSend := neverListen
  patchSend := neverListen

  // Most recent data to be sent to each channel (i.e. off the other channel)
  internalData := reflect.ValueOf(nil)
  patchData := reflect.ValueOf(nil)

  // Indicates whether the data for the patch channel has come of internal or
  // is still the default
  sendOnPatch := false

  pType := BuildPortType(c, t, broadcast)

	port := Port{portType: pType, setPatchChan: setPatchChan, stop: stopChan}

  // Static cases 
  stopCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: stop}
  setCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: setPatch}

	go func() {
    for {
      internalRecvCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: internalRecv}
      patchRecvCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: patchRecv}
      internalSendCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: internalSend, Send: internalData}
      patchSendCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: patchSend, Send: patchData}

      selCases := []reflect.SelectCase{stopCase, setCase, internalRecvCase, patchRecvCase,
        internalSendCase, patchSendCase}

      chosen, recv, recvOK := reflect.Select(selCases)

      switch chosen {
      case 0: // Stop
        return
      case 1: // New patch output
        if recvOK {
          newPatch := recv
          if !newPatch.IsNil() {
            if newPatch.Kind() != reflect.Chan {
              setPatchChan <- errors.New("Patch channel is not a chan")
            }

            dir := newPatch.Type().ChanDir()

            if dir != reflect.BothDir && dir == internal.Type().ChanDir() {
              setPatchChan <- errors.New("Patch channel is the incorrect direction for this port")
            }

            newType := newPatch.Type().Elem()
            intType := internal.Type().Elem()
            if !(newType.AssignableTo(intType) && intType.AssignableTo(intType)) {
              setPatchChan <- errors.New("Patch channel type is not compatible with this port")
            }

            patch = newPatch

            if dir != reflect.SendDir {
              patchRecvDefault = patch
              patchRecv = patch
            } else {
              patchRecv = reflect.ValueOf(nil)
            }

            if dir != reflect.RecvDir && sendOnPatch {
              patchSend = patch
            } else {
              patchSend = neverListen
            }
          } else {
            patch = newPatch
            patchSend = neverListen
            patchRecv = newPatch
            setPatchChan <- nil
          }
        }
      case 2: // Recieved data from internal
        // Ignore it if there i
        if !patch.IsNil() && patch.Type().ChanDir() != reflect.RecvDir {
          if recvOK {
            patchData = recv
            patchSend = patch
            // Prevent "feedback" by disabling listening till data arrives
            patchRecv = reflect.ValueOf(nil)
          } else {
            if !patch.IsNil() {
              internal.Close()
              return
            }
          }
        }
      case 3: // Recieved data from patch
        // Only do anything if we expect to send data on internal
        if internal.Type().ChanDir() != reflect.RecvDir {
          if recvOK {
            internalData = recv
            internalSend = internal
            internalRecv = reflect.ValueOf(nil)
          } else {
            patch.Close()
            return
          }
        }
      case 4: // Sent data to internal
        internalSend = neverListen
        internalRecv = internalRecvDefault
      case 5: // Sent data to patch
        patchSend = neverListen
        patchRecv = patchRecvDefault
      }
		}
	}()

  return port, nil
}

func (p Port) Type() Type {
	return p.portType
}

func (p Port) ObjectMethod(name string) Function {
	return nil
}

func (p Port) SetPatchChan(c interface {}) error {
	p.setPatchChan <- c
  err := <-p.setPatchChan
  return err.(error)
}

func (p Port) IsInput() bool {
  return p.portType.IsInput()
}

func (p Port) IsOutput() bool {
  return p.portType.IsOutput()
}

// Port Type

func BuildPortType(c interface {}, t Type, broadcast bool) PortType {
  input := reflect.TypeOf(c).ChanDir() != reflect.SendDir
  output := reflect.TypeOf(c).ChanDir() != reflect.RecvDir

  return PortType{valType: t, broadcast: broadcast, input: input, output: output}
}

func (pt PortType) Name() string {
	return "Port[" + pt.valType.Name() + "]"
}

func (PortType) Method(name string) Method {
	return nil
}

func (PortType) Methods() []string {
	return []string{}
}

func (pt PortType) IsInput() bool {
  return pt.input
}

func (pt PortType) IsOutput() bool {
  return pt.output
}

// Convenience types for embedding in structs
type InputPort Port
type OutputPort *Port

func (ip InputPort) BuildInput(c chan<- Object, t Type, broadcast bool) {
  port, _ := NewPort(c, t, broadcast)
  ip = InputPort(port)
}

func (ip InputPort) Input() Port {
  return Port(ip)
}
