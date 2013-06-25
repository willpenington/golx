package patch

// Core direct patching functionallity

import (
  "reflect"
  "errors"
  "fmt"
)


// Internal representation of a patch

type patchData struct {
  inputDelegator reflect.Value
  input reflect.Value
  inputChan reflect.Value

  outputDelegator reflect.Value
  output reflect.Value
  outputChan reflect.Value

  commonChan reflect.Value

  stop chan bool
}

// Request structures to send data over a channel

type patchRequest struct {
  Output interface {}
  OutputDelegator interface {}

  Input interface {}
  InputDelegator interface {}
}

type unpatchRequest struct {
  Output interface {}
  Input interface {}
}

// Information about the current patching. These variable should only be used
// by functions called from the patch manager to keep the information consistent

// Patches by the objects they link
var inputPatches map[reflect.Value] map[*patchData] bool
var outputPatches map[reflect.Value] map[*patchData] bool


// Patches by the channels they use
var inputChanPatches map[reflect.Value] *patchData
var outputChanPatches map[reflect.Value] *patchData

// Channels for communicating with the patch manager. These should only be used
// by the patch manager and the function matching the request type
var newPatchRequest chan patchRequest 
var newPatchError chan error

var newUnpatchRequest chan unpatchRequest
var newUnpatchError chan error

// Set up data, channels and start the patch manager
func init() {
  inputPatches = make(map[reflect.Value] (map[*patchData] bool))
  outputPatches = make(map[reflect.Value] (map[*patchData] bool))

  inputChanPatches = make(map[reflect.Value] *patchData)
  outputChanPatches = make(map[reflect.Value] *patchData)

  newPatchRequest = make(chan patchRequest)
  newPatchError = make(chan error)

  newUnpatchRequest = make(chan unpatchRequest)
  newUnpatchError = make(chan error)

  go patchManager()
}

// Indicates the direction of the attempted patch for shared methods
type patchDirection int

const (
  patchInputDir patchDirection = iota
  patchOutputDir
)

// Name of the method that should be called to get the channel for this direction
func (dir patchDirection) GetMethod() string {
  switch dir {
  case patchInputDir:
    return "Input"
  case patchOutputDir:
    return "Output"
  default:
    panic("Unexpected patch direction")
  }
}

// Name of this direction for errors and logging
func (dir patchDirection) Label() string {
  switch dir {
  case patchInputDir:
    return "Input"
  case patchOutputDir:
    return "Output"
  default:
    panic("Unexpected patch direction")
  }
}

// Name of the method that should be called to get the delegate/list of delegates
// for this direction
func (dir patchDirection) DelegateMethod() string {
  return "Delegate" + dir.GetMethod()
}

// Name of the method that should be called to set the channel for this direction
func (dir patchDirection) SetMethod() string {
  return "Set" + dir.GetMethod()
}

// Returns the type the object uses for direct patching
func patchType(obj reflect.Value, dir patchDirection) (reflect.Type, error) {
  // Try finding a type from the set channel method first as it is prefered
  setChanMethod, err := setChanMethod(obj, dir)

  if err == nil {
    return setChanMethod.Type().In(0), nil
  }

  // If that fails try the get channel method
  getChanMethod, err := getChanMethod(obj, dir)

  if err == nil {
    return getChanMethod.Type().Out(0), nil
  }

  // If neither works the object doesn't implement the correct interface and 
  // doesn't support patching
  return reflect.TypeOf(nil), errors.New(dir.Label() + " does not support direct patching")
}

// Returns the method to get the channel for direct patching
func getChanMethod(obj reflect.Value, dir patchDirection) (reflect.Value, error) {
  getChanMethod := obj.MethodByName(dir.GetMethod())

  // Check the method actually exists
  if !getChanMethod.IsValid() {
    return reflect.ValueOf(nil), errors.New("Object does not implement " + dir.Label() + " patching")
  }

  // Check it doesn't require arguments as we can't generate them
  if getChanMethod.Type().NumIn() != 0 {
    return reflect.ValueOf(nil), errors.New(dir.GetMethod() + " requires arguments")
  }

  // Check it gives at least one output. We can ignore the others
  if getChanMethod.Type().NumOut() < 1 {
    return reflect.ValueOf(nil), errors.New(dir.GetMethod() + " does not return any value")
  }

  // Check that the first output is a channel, otherwise it doesn't help us
  retType := getChanMethod.Type().Out(0)
  if retType.Kind() != reflect.Chan {
    return reflect.ValueOf(nil), errors.New(dir.GetMethod() + " does not return a channel")
  }

  return getChanMethod, nil
}

// Returns the method to set the channel for direct patching
func setChanMethod(obj reflect.Value, dir patchDirection) (reflect.Value, error) {
  fmt.Println("Looking for set method")
  method := obj.MethodByName(dir.SetMethod())

  fmt.Println("Checking method is valid")
  // Check the method actually exists
  if !method.IsValid() {
    return reflect.ValueOf(nil), errors.New("Object does not implement " + dir.Label() + " patching")
  }

  fmt.Println("Looking at arguments")
  // Check the method takes exactly one argument. We don't know what to pass if there are others
  if method.Type().NumIn() != 1 {
    return reflect.ValueOf(nil), errors.New(dir.SetMethod() + " requires wrong number of arguments")
  }

  fmt.Println("Getting argument type")
  // Check the type of the first argument is a channel, because that's what were going to try and give it
  argType := method.Type().In(0)
  fmt.Println("Checking argument type")
  if argType.Kind() != reflect.Chan {
    return reflect.ValueOf(nil), errors.New(dir.SetMethod() + " does not take a channel as its argument")
  }

  fmt.Println("Done in setChanMethod")
  return method, nil
}

func canPatch(output, input interface {}) (bool, error) {
  return canPatchVal(reflect.ValueOf(output), reflect.ValueOf(input))
}

// Check if an input and output can be directly patched together
func canPatchVal(output, input reflect.Value) (bool, error) {
  // Get the type of the output
  outType, outErr := patchType(output, patchOutputDir)

  // Rely on patchType to provide an error if it doesn't match the interface
  if outErr != nil {
    return false, errors.New("Output does not support patching")
  }

  // Repeat for the input
  inType, inErr := patchType(input, patchInputDir)

  if inErr != nil {
    return false, errors.New("Input does not support patching")
  }

  // Check the output is compatible with the input
  /* This check is stricter than the one used by the actual patch system but
     properly replicating it's check involves knowing the input and output both
     use channel getters rather than setters. This check passes most things the
     real one does, nothing it doesn't and is simpler to implement.
  */
  if !(outType.AssignableTo(inType)) {
    return false, errors.New("Input and output are not compatible")
  }

  return true, nil
}

// Copys values from output onto input till a value is recieved on stop
// output and input should be compatible channels
// closing output closes input and quits
func proxy(output, input reflect.Value, stop chan bool) {
  // Build reflection select structure
  stopVal := reflect.ValueOf(stop)

  recvCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: output }
  stopCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: stopVal }

  recvSelect := []reflect.SelectCase{recvCase, stopCase}

  for {
    // Listen from data from output
    chosen, recv, recvOK := reflect.Select(recvSelect)

    switch chosen {
    case 0:
      if recvOK {
        // Convert the data to the type expected by the input
        data := recv.Convert(input.Type().Elem())

        // Build a select case for sending the data
        sendCase := reflect.SelectCase{Dir: reflect.SelectSend, Chan: input, Send: data}
        sendSelect := []reflect.SelectCase{sendCase, stopCase}

        // Send the data or wait for a quit
        sendChosen, _, _ := reflect.Select(sendSelect)

        switch sendChosen {
        case 0:
          // Data sent sucessfully
        case 1:
          // Stop recieved
          return
        }
      } else {
        // Close the input if the output is closed
        input.Close()

        // Listen for stop so to avoid blocking goroutines that don't know
        // the proxy has stopped. The risk of leaking goroutines is lower than
        // the risk of randomly blocking other functions that try and send a
        // stop to the goroutine.
        _ = <-stop

        return
      }
    case 1:
      // Stop recieved
      return
    }
  }
}

// Convert a patch request to a patchData struct
// Prevents pointers to patchData being anywhere they
// don't need to and hides a bit of reflection stuff
func (req patchRequest) buildPatch() *patchData {
  patch := new(patchData)

  patch.output = reflect.ValueOf(req.Output)
  patch.outputDelegator = reflect.ValueOf(req.OutputDelegator)

  patch.input = reflect.ValueOf(req.Input)
  patch.inputDelegator = reflect.ValueOf(req.InputDelegator)

  return patch
}

// Build, configure and store a new patch
func addPatch(req patchRequest) error {
  fmt.Println("Building patch data")
  // Convert the request to a patchData struct
  patch := req.buildPatch()

  fmt.Println("Getting accessor methods")
  // Check the output and input are patchable and find the functions for dealing with them
  setOutputChan, setOutputChanErr := setChanMethod(patch.output, patchOutputDir)
  getOutputChan, getOutputChanErr := getChanMethod(patch.output, patchOutputDir)
  setInputChan, setInputChanErr := setChanMethod(patch.input, patchInputDir)
  getInputChan, getInputChanErr := getChanMethod(patch.input, patchInputDir)
  fmt.Println("Retrieved accessors")

  // Check the output supports at least one method of patching
  if setOutputChanErr != nil && getOutputChanErr != nil {
    return errors.New("Output does not support direct patching")
  }

  // Check the input supports at least one method of patching
  if setInputChanErr != nil && getInputChanErr != nil {
    return errors.New("Input does not support direct patching")
  }

  // Flags indicate if the patch or object is supplying the channel
  // This is the same information as the errors but in a more readable form
  setOutput := false
  setInput := false

  if setOutputChanErr == nil {
    setOutput = true
  }

  if setInputChanErr == nil {
    setInput = true
  }

  fmt.Println("Checking flags")
  // Get the types of the channels being patched
  var outType reflect.Type
  var inType reflect.Type

  if setOutput {
    outType = setOutputChan.Type().In(0)
  } else {
    outType = getOutputChan.Type().Out(0)
  }

  if setInput {
    inType = setInputChan.Type().In(0)
  } else {
    inType = getInputChan.Type().Out(0)
  }

  // getChanMethod/setChanMethod garuntee the type is a channel so no check is needed

  // Check that the types are compatible
  if (setInput || setOutput) && !(outType.AssignableTo(inType)) {
    return errors.New("Output is not compatible with input")
  } else if !(outType.Elem().ConvertibleTo(inType.Elem())) {
    // If both ports are providing channels we can be less strict and convert in the
    // proxy
    return errors.New("Output is not compatible with input")
  }

  fmt.Println("Loading existing data")
  // Load the existing data
  inputPatchList, inputPatchListExists := inputPatches[patch.input]
  outputPatchList, outputPatchListExists := outputPatches[patch.output]

  // Build new lists if they don't exist yet
  if !outputPatchListExists {
    outputPatchList = make(map[*patchData] bool)
  }

  if !inputPatchListExists {
    inputPatchList = make(map[*patchData] bool)
  }

  // Look for existing patches 
  for inPatch, _ := range inputPatchList {
    if inPatch.output == patch.output {
      return errors.New("Output and Input are already patched")
    }
  }

  // Add the new patch to the lists
  outputPatchList[patch] = true
  inputPatchList[patch] = true

  var inputDelegatorPatchList map[*patchData] bool
  var inputDelegatorPatchListExists bool
  if !patch.inputDelegator.IsNil() {
    inputDelegatorPatchList, inputDelegatorPatchListExists = inputPatches[patch.inputDelegator]

    if !inputDelegatorPatchListExists {
      inputDelegatorPatchList = make(map[*patchData] bool)
    }

    for inPatch, _ := range inputDelegatorPatchList {
      if inPatch.output == patch.output {
        return errors.New("Output and Input are already patched")
      }
    }

    inputDelegatorPatchList[patch] = true
  }

  var outputDelegatorPatchList map[*patchData] bool
  var outputDelegatorPatchListExists bool
  if !patch.outputDelegator.IsNil() {
    outputDelegatorPatchList, outputDelegatorPatchListExists = outputPatches[patch.outputDelegator]

    if !outputDelegatorPatchListExists {
      outputDelegatorPatchList = make(map[*patchData] bool)
    }

    outputDelegatorPatchList[patch] = true
  }

  fmt.Println("Building common channel")
  // If the patch is supplying the channel to both sides build a common channel using the type
  // of the output
  if setOutput && setInput {
    patch.commonChan = reflect.MakeChan(outType, 0)
  }

  fmt.Println("Getting existing channels")
  if !setOutput {
    // Get the channel from output
    patch.outputChan = getOutputChan.Call([]reflect.Value{})[0]

    // Store it as the common channel
    patch.commonChan = patch.outputChan

    // Ensure the channel isn't already in the patch
    _, outputChanExists := outputChanPatches[patch.outputChan]
    if outputChanExists {
      return errors.New("Output channel is already patched")
    }
  }

  if !setInput {
    // Get the channel from input
    patch.inputChan = getInputChan.Call([]reflect.Value{})[0]

    // Store it as the common channel

    // Ensure the channel isn't already in the patch
    _, inputChanExists := inputChanPatches[patch.inputChan]
    if inputChanExists {
      return errors.New("Input channel is already patched")
    }
  }

  fmt.Println("Setting channels")
  // Set the channels as nescessary
  if setOutput {
    patch.outputChan = patch.commonChan
    setOutputChan.Call([]reflect.Value{patch.commonChan})
  }

  if setInput {
    patch.inputChan = patch.commonChan
    setInputChan.Call([]reflect.Value{patch.commonChan})
  }

  // Set the default value of stop. It should only be a channel if there is a proxy
  patch.stop = nil

  fmt.Println("Starting proxy")
  // If both ports are providing channels proxy them together, build the stop
  // channel for the proxy and unset the common channel
  if !(setOutput || setInput) {
    patch.stop = make(chan bool)
    proxy(patch.outputChan, patch.inputChan, patch.stop)
  }

  fmt.Println("Storing changes to data")
  outputPatches[patch.output] = outputPatchList
  if !patch.outputDelegator.IsNil() {
    outputPatches[patch.outputDelegator] = outputDelegatorPatchList
  }

  inputPatches[patch.input] = inputPatchList
  if !patch.inputDelegator.IsNil() {
    inputPatches[patch.inputDelegator] = inputDelegatorPatchList
  }

  outputChanPatches[patch.outputChan] = patch
  inputChanPatches[patch.inputChan] = patch

  return nil
}

func removePatch(req unpatchRequest) error {
  input := reflect.ValueOf(req.Input)
  output := reflect.ValueOf(req.Output)

  var patch *patchData = nil

  inputPatchList, inputPatchListExists := inputPatches[input]

  if !inputPatchListExists {
    return errors.New("Output and input are not patched")
  }

  for inPatch, _ := range inputPatchList {
    if inPatch.output == output || inPatch.outputDelegator == output {
      patch = inPatch
      break
    }
  }

  if patch == nil {
    return errors.New("Output and input are not patched")
  }

  // Stop the proxy if it exists
  if patch.stop != nil {
    patch.stop <- true
  }

  // Get channel setter methods
  setOutputChan, setOutputChanErr := setChanMethod(patch.output, patchOutputDir)
  setInputChan, setInputChanErr := setChanMethod(patch.input, patchInputDir)

  // Set the channels to nil if nescessary
  if setOutputChanErr != nil {
    setOutputChan.Call([]reflect.Value{reflect.ValueOf(nil)})
  }

  if setInputChanErr != nil {
    setInputChan.Call([]reflect.Value{reflect.ValueOf(nil)})
  }

  // Remove the record of the patch
  delete(outputChanPatches, patch.outputChan)
  delete(inputChanPatches, patch.inputChan)

  delete(outputPatches[patch.output], patch)
  delete(inputPatches[patch.input], patch)

  if !patch.outputDelegator.IsNil() {
    delete(outputPatches[patch.outputDelegator], patch)
  }

  if !patch.inputDelegator.IsNil() {
    delete(inputPatches[patch.inputDelegator], patch)
  }

  return nil
}

func directPatch(output, input, outputDelegator, inputDelegator interface {}) error {
  fmt.Println("Adding new direct patch")
  if output == nil {
    return errors.New("Output must not be nil")
  }

  if input == nil {
    return errors.New("Input must not be nil")
  }

  fmt.Println("Building new request")
  req := patchRequest{output, outputDelegator, input, inputDelegator}

  fmt.Println("Sending request")
  newPatchRequest <- req
  fmt.Println("Reading errors")
  err := <-newPatchError

  fmt.Println("Done")
  return err
}

func Unpatch(output, input interface {}) error {
  newUnpatchRequest <- unpatchRequest{output, input}
  err := <-newUnpatchError

  return err
}

func Patch(output, input interface {}) error {
  return directPatch(output, input, nil, nil)
}


func patchManager() {
  for {
    fmt.Println("Waiting for request")
    select {
    case patchReq := <-newPatchRequest:
      fmt.Println("Got patch request")
      err := addPatch(patchReq)
      fmt.Println("Sending error")
      newPatchError <- err
    case unpatchReq := <-newUnpatchRequest:
      fmt.Println("Got unpatch request")
      err := removePatch(unpatchReq)
      fmt.Println("Sending error")
      newUnpatchError <- err
    }
  }
}
