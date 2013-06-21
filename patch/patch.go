/*
Provides support for patching together various goroutines
*/
package patch

import (
  "reflect"
  "errors"
)

type patch struct {
  input InputPort
  output OutputPort
  quit chan bool
}

type InputPort interface {
  InputChannel() chan interface {}
}

type OutputPort interface {
  OutputChannel() chan interface {}
}

var inputPatches map[InputPort] *patch
var outputPatches map[OutputPort] *patch

func init() {
  inputPatches = make(map[InputPort] *patch)
  outputPatches = make(map[OutputPort] *patch)
}

func Patch(output OutputPort, input InputPort) error {

  _, inExists := inputPatches[input]
  _, outExists := outputPatches[output]

  if inExists {
    return errors.New("Input is already patched")
  }

  if outExists {
    return errors.New("Output is already patched")
  }

  inChannel := reflect.ValueOf(input.InputChannel())
  outChannel := reflect.ValueOf(output.OutputChannel())

  inType := inChannel.Type().Elem()
  outType := outChannel.Type().Elem()

  if !(outType.AssignableTo(inType)) {
    return errors.New("Input and Output are not compatible")
  }

  p := new(patch)

  p.input = input
  p.output = output
  p.quit = make(chan bool)

  inputPatches[input] = p
  outputPatches[output] = p

  recvCase := new(reflect.SelectCase)
  recvCase.Chan = outChannel
  recvCase.Dir = SelectRecv

  quitCase := new(reflect.SelectCase)
  quitCase.Chan = reflect.ValueOf(quit)
  quitCase.Dir = SelectRecv

  listenSelect := make([]reflect.SelectCase)
  append(listenSelect, recvCase)
  append(listenSelect, quitCase)

  go func() {
    for {
      chosen, val, ok := reflect.Select(listenSelect)

      if chosen == 0 {
        // Recieved from the output port
        sendVal := val.ConvertTo(outType)
        inChannel.Send(sendVal)
      } else if chosen == 1 {
        // Recieved a quit signal
        return
      }
    }
  }()

  return nil
}

func Unpatch(output OutputPort, input InputPort) error {
  inPatch, inExists := inputPatches[input]
  outPatch, outExists := outputPatches[output]

  if !inExists || !outExists || inPatch != outPatch {
    return errors.New("Patch does not exist")
  }

  inPatch.quit <- true
  delete(inputPatches, input)
  delete(outputPatches, output)

  return nil
}
