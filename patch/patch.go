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

type InputPort {
  InputChannel() interface {}
}

type OutputPort {
  OutputChannel() interface {}
}

var inputPatches map[InputPort] *patch
var outputPatches map[OutputPort] *patch

func Patch(output OutputPort, input InputPort) error {

  _, inExists := inputPatches[input]
  _, outExists := outputPatches[output]

  if inExists {
    return errors.New("Input is already patched")
  }

  if outExists {
    return errors.New("Output is already patched")
  }

  inChannel := input.InputChannel()
  outChannel := output.OutputChannel()

  inType := reflect.TypeOf(inChannel)
  outType := reflect.TypeOf(outChannel)

  if !(outType.AssignableTo(inType)) {
    return errors.New("Input and Output are not compatible")
  }

  p := make(patch)

  p.input = input
  p.output = output
  p.quit = make(chan bool)

  inputPatches[input] = p
  outputPatches[output] = p

  go func() {
    for {
      select {
        case val := <-outChannel
          inChannel <- val
        case _ = <-p.quit
          return
      }
    }
  }

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
