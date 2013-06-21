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
}

type OutputPort interface {
}

var inputPatches map[InputPort] *patch
var outputPatches map[OutputPort] *patch

func init() {
  inputPatches = make(map[InputPort] *patch)
  outputPatches = make(map[OutputPort] *patch)
}

func getChan(i interface {}, method string) (reflect.Value, error) {
  val := reflect.ValueOf(i)

  m := val.MethodByName(method)

  if !m.IsValid() {
    return reflect.ValueOf(nil), errors.New("does not support " + method + " method.")
  }

  if m.Type().NumIn() != 0 {
    return reflect.ValueOf(nil), errors.New("requires incorrect number of arguments for " + method + "method.")
  }

  if m.Type().NumOut() != 1 {
    return reflect.ValueOf(nil), errors.New("returns too many values for " + method + "method.")
  }

  retType := m.Type().Out(0)

  if retType.Kind() != reflect.Chan {
    return reflect.ValueOf(nil), errors.New("does not return a channel from " + method + "method.")
  }

  args := make([]reflect.Value, 0)
  result := m.Call(args)
  return result[0], nil
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

  inChannel, err := getChan(input, "InputChannel")

  if err != nil {
    return errors.New("Input " + err.Error())
  }

  outChannel, err := getChan(output, "OutputChannel")

  if err != nil {
    return errors.New("Output " + err.Error())
  }

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

  listenSelect := make([]reflect.SelectCase, 2)

  listenSelect[0].Chan = outChannel
  listenSelect[0].Dir = reflect.SelectRecv

  listenSelect[1].Chan = reflect.ValueOf(p.quit)
  listenSelect[1].Dir = reflect.SelectRecv

  go func() {
    for {
      chosen, val, _ := reflect.Select(listenSelect)

      if chosen == 0 {
        // Recieved from the output port
        sendVal := val.Convert(outType)
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
