package chanutil

import (
  "reflect"
  "errors"
)

func DeliverWhenPossible(input, output interface {}) error {
  inputVal := reflect.ValueOf(input)
  outputVal := reflect.ValueOf(output)

  if inputVal.Kind() != reflect.Chan {
    return errors.New("Input must be a channel")
  }

  if outputVal.Kind() != reflect.Chan {
    return errors.New("Output must be a channel")
  }

  if !(inputVal.Type().Elem().ConvertibleTo(outputVal.Type().Elem())) {
    return errors.New("Input is not compatible with output")
  }

  go func() {
    dataAvailable := false
    data := reflect.ValueOf(nil)

    var selCases []reflect.SelectCase
    recvCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: inputVal}

    for {

      if dataAvailable {
        sendCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: outputVal, Send: data}
        selCases = []reflect.SelectCase{recvCase, sendCase}
      } else {
        selCases = []reflect.SelectCase{recvCase}
      }

      chosen, recv, recvOK := reflect.Select(selCases)

      switch chosen {
      case 0:
        if recvOK {
          data = recv
          dataAvailable = true
        } else {
          return
        }
      case 1:
        dataAvailable = false
      }
    }
  }()

  return nil
}
