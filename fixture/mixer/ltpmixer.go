package mixer

import (
  "reflect"
  "errors"
  "fmt"
  "golx/patch/chanutil"
)

/*
A 
*/
type LTPMixer struct {
  mixType reflect.Type
  input reflect.Value
  stop reflect.Value
  defaultVal reflect.Value

  isDefault bool
  notifyDefault chan bool
  notifyDefaultPublic chan bool
}

type ltpLink struct {
  Data interface {}
  Stop interface {}
  Replace interface {}
  Final bool
}

func NewLTPMixer(output, def interface {}) (*LTPMixer, error) {
  mixer := new(LTPMixer)

  outputVal := reflect.ValueOf(output)
  mixer.defaultVal = reflect.ValueOf(def)

  if outputVal.Kind() != reflect.Chan {
    return nil, errors.New("Output must be a channel")
  }

  mixer.mixType = outputVal.Type().Elem()

  if mixer.defaultVal.Type().ConvertibleTo(mixer.mixType) == false {
    return nil, errors.New("Default value must match output type")
  }

  inputType := reflect.ChanOf(reflect.BothDir, reflect.ChanOf(reflect.BothDir, mixer.mixType))
  mixer.input = reflect.MakeChan(inputType, 0)

  mixer.stop = reflect.ValueOf(make(chan bool))

  mixer.isDefault = true
  mixer.notifyDefault = make(chan bool)
  mixer.notifyDefaultPublic = make(chan bool)

  chanutil.DeliverWhenPossible(mixer.notifyDefault, mixer.notifyDefaultPublic)

  go mixer.mix(outputVal)

  return mixer, nil
}

var counter int

func init() {
  counter = 1
}

func token() int {
  counter += 1
  return counter
}

func (mixer *LTPMixer) Stop() {
  mixer.stop.Send(reflect.ValueOf(true))
}

// Inner loop of the mixer.
func (mixer *LTPMixer) mix(output reflect.Value) {
  head := mixer.finalLink()

  loopToken := token()

  for {
    headData := head.Elem().FieldByName("Data").Elem()
    headReplace := head.Elem().FieldByName("Replace").Elem()

    selCases := make([]reflect.SelectCase, 4)

    selCases[0] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: mixer.input}
    selCases[1] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: headData }
    selCases[2] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: headReplace}
    selCases[3] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: mixer.stop}

    // selCases := []reflect.SelectCase{inputCase, dataCase, replaceCase, stopCase}

    fmt.Println(loopToken, "In main loop")

    chosen, recv, _ := reflect.Select(selCases)

    switch chosen {
    case 0:
      // Add a new channel to the list
      fmt.Println(loopToken, "Adding a new input to the list")
      head = addLink(recv, head)
    case 1:
      fmt.Println(loopToken, "Got data from top of the list")
      output.Send(recv.Convert(mixer.mixType))

      // Notify any listners if the previous value sent was the default
      if mixer.isDefault {
        mixer.isDefault = false
        mixer.notifyDefault <- false
      }

    case 2:
      fmt.Println(loopToken, "Top of list quitting")
      head = recv
      if reflect.Indirect(recv).FieldByName("Final").Bool() == true {
        output.Send(mixer.defaultVal)
        mixer.notifyDefault <- true
      }
    case 3:
      fmt.Println(loopToken, "Got stop. Quitting")
      reflect.Indirect(head).FieldByName("Stop").Send(recv)
      return
    default:
      fmt.Println(loopToken, "Got unexpected index")
    }

  }
}

func addLink(input, tail reflect.Value) reflect.Value {
  head := new(ltpLink)
  head.Final = false // This is not the final link in the chain

  // Assume that input is a channel of the correct type
  // Build a new channel type to avoid accidently copying channel direction
  outType := reflect.ChanOf(reflect.BothDir, input.Type().Elem())
  dataVal := reflect.MakeChan(outType, 0)
  head.Data = dataVal.Interface()

  replace := make(chan *ltpLink)
  replaceVal := reflect.ValueOf(replace)
  head.Replace = replace
  stop := make(chan bool)
  stopVal := reflect.ValueOf(stop)
  head.Stop = stop

  go func() {
    currentVal := reflect.ValueOf(nil)
    dataAvailable := false
    dataSeen := false

    tkn := token()

    for {
      fmt.Println(tkn, "In link loop")
      fmt.Println(tkn, "dataAvailable: ", dataAvailable)
      fmt.Println(tkn, "dataSeen: ", dataSeen)

      tailReplace := tail.Elem().FieldByName("Replace").Elem()

      // Build select
      recvCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: input}
      stopCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: stopVal}
      tailCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: tailReplace}

      selCases := []reflect.SelectCase{recvCase, stopCase, tailCase}

      // Set the indicies of the optional cases to things Select will never return
      // so their handlers only run if they are added
      if dataAvailable {
        sendCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: dataVal, Send: currentVal}
        selCases = append(selCases, sendCase)
      }

      if !dataSeen {
        tailData := tail.Elem().FieldByName("Data").Elem()
        forwardCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: tailData}
        selCases = append(selCases, forwardCase)
      }

      chosen, recv, recvOK := reflect.Select(selCases)

      // Hide the missing option if no attempt was made to send data
      if !dataAvailable && chosen >= 3 {
        chosen += 1
      }

      switch chosen {
      case 0:
        if recvOK {
          // New data from input
          fmt.Println(tkn, "Got data")
          dataSeen = true
          dataAvailable = true
          currentVal = recv
        } else {
          // Input channel closed
          fmt.Println(tkn, "Quitting")
          replaceVal.Send(tail)
          return
        }
      case 1:
        // Stop command
        reflect.Indirect(tail).FieldByName("Stop").Send(reflect.ValueOf(true))
        return
      case 2:
        // Next link quit, replace tail
        if recvOK {
          fmt.Println(tkn, "Child quit")
          tail = recv
        }
      case 3:
        fmt.Println(tkn, "Sent data")
        dataAvailable = false
      case 4:
        if recvOK {
          fmt.Println(tkn, "Got child's data")
          dataAvailable = true
          currentVal = recv
        }
      }
    }
  }()

  return reflect.ValueOf(head)
}

func (mixer *LTPMixer) finalLink() reflect.Value {
  link := new(ltpLink)
  link.Replace = make(chan *ltpLink)
  link.Final = true

  dataType := reflect.ChanOf(reflect.BothDir, mixer.mixType)
  link.Data = reflect.MakeChan(dataType, 0).Interface()

  stop := make(chan bool)
  go func() { _ = <-stop }()

  link.Stop = stop

  return reflect.ValueOf(link)
}

func (mixer *LTPMixer) AddInput(input interface {}) error {
  inputVal := reflect.ValueOf(input)

  if inputVal.Kind() != reflect.Chan {
    return errors.New("Input must be a channel")
  }

  if inputVal.Type().Elem().ConvertibleTo(mixer.mixType) == false {
    return errors.New("Mixer type does not match input type")
  }

  mixer.input.Send(inputVal)

  return nil
}
