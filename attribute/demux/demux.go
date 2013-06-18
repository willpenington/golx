/*
HTP/LTP and Priority based channel demuxer

The system consists of two main parts: the Value Mixer and the Priority
Selector (implemented as the Demuxer)

Mixer:

The HTP/LTP section uses a structure similar to a linked list in channels and
go routines that listens to each input channel and when passed a channel will
pass the most recent value off each input (if it has recieved one) into the
channel in the reverse order to how the channels were added (i.e. most recent
channel first). This is implemented with the data channel in the chainLink
structure. The goroutines also check if their input channel has been closed and
automatically remove themseleves from the list if it has (i.e. releasing
control). This is implemented with the replace channel in the chainLink
structure which points in the opposite direction to data. To remove itself from
the chain a goroutine passes its tail to the previous goroutine which replaces
its own tail with the data it recieves.

The actual selection is performed by a function that takes a channel of values
and reduces it to a single value. The channel always recieves at least one
value as the selector is not used if there are no values in the channel.

Priority Selector:
This is significantly simpler than the mixer and currently uses a very naive
and not particularly thread safe implentation. It maintains a simple map
from integer priorities to mixers and returns the value of the highest priority
mixer with a value. If no mixer can provide a value it returns a default.

*/
package demux

import (
	"errors"
	"math"
  "golx/attribute"
)

/*
MixerValue contains both the actual value and the context that sent it (so the
numbers can be exciting colours)
*/
type MixerValue struct {
  Value attribute.AttrValue
  //TODO Add source type
}

// Priorities are really just integers, but this keeps the constants seperate
// int32 is used to make it easier to know the lowest value
type Priority int32

type Demuxer struct {
	defaultVal MixerValue
	// Mixers are stored in a map to garuntee no two will have the same priority
	mixers map[Priority]Mixer
}

// TODO Replace struct data with channels and a goroutine

func NewDemuxer(defaultVal MixerValue) *Demuxer {
	demuxer := new(Demuxer)
	demuxer.mixers = make(map[Priority]Mixer)
	demuxer.defaultVal = defaultVal
	return demuxer
}

func (demuxer *Demuxer) Value() MixerValue {
	// Slightly hacky, temporary solution to finding the highest priority value
	max := Priority(math.MinInt32)
	val := demuxer.defaultVal

	for priority, mixer := range demuxer.mixers {
		if priority > max {
			tempVal, valid := demuxer.mixers[priority]
			if valid {
				max = priority
				val = tempVal
			}
		}
	}

	return val

}

func (demuxer *Demuxer) SetDefault(defValue MixerValue) {
	demuxer.defaultVal = defValue
}

func (demuxer *Demuxer) AddMixer(mixer *Mixer, priority Priority) error {
	_, exists := demuxer.mixers[priority]
	if exists {
		return errors.New("Mixer already exists with this priority")
	}

	demuxer.mixers[priority] = mixer
}

func (demuxer *Demuxer) DeleteMixer(priority Priority) {
	delete(demuxer.mixers, priority)
}

// TODO implement managment of mixers that have been added

/* Mixer stuff */

type Mixer struct {
	data      chan chan MixerValue
	newInputs chan chan MixerValue
	Selector  func(chan MixerValue) MixerValue
}

type chainLink struct {
	/*
	  links respond to messages on the data channel by sending their current value
	  on the channel they recieve
	*/
	data chan chan MixerValue
	/*
	  links respond to messages on the replace channel by replacing their tail with
	  the value they recieve
	*/
	replace chan *chainLink
}

// Initialise the mixer and start its background goroutine
func NewMixer(selector func(chan MixerValue) MixerValue) *Mixer {
	mixer := new(Mixer)
	mixer.data = make(chan chan MixerValue)
	mixer.newInputs = make(chan chan MixerValue)
	mixer.Selector = selector

	// This link represents the end of the chain (which is empty at the start)
	// No other link should have nil channels
	head := new(chainLink)
	head.data = nil
	head.replace = nil

	go func() {
		for {
			select {
			case input := <-mixer.newInputs:
				// New channels are added at the start of the chain
				head = addLink(input, head)
			case dataChan := <-mixer.data:
				// Channels passed into the list are double handled so head can be
				// hidden in the goroutine and swapped
				if head.data != nil {
					head.data <- dataChan
				} else {
					// closing the channel represents the end of the chain
					close(dataChan)
				}
      case newHead := <-head.replace
        head = newHead
			}
		}
	}()

	return mixer
}

// Helper function to build chainlink structs
func newChainLink() *chainLink {
	link := new(chainLink)
	link.data = make(chan (chan MixerValue))
	link.replace = make(chan *chainLink)

	return link
}

// Adds new inputs to the chain, hiding the internal use of channels
func (mixer *Mixer) AddInput(input chan MixerValue) {
	mixer.newInputs <- input
}

// Calculates the current value of the mixer
func (mixer *Mixer) Value() (MixerValue, bool) {
	// Pass a channel into the list to get the values in order
	collector := make(chan MixerValue)
	go func() { mixer.data <- collector }()

	// Checking if the channel is empty damages the list, so collector must be
	// reassigned
	hasVal, collector := hasValues(collector)

	if hasVal {
		// Use the selector only on channels that aren't empty
		// Otherwise the program may hang
		return mixer.Selector(collector), true
	}

	// If the channel is empty the mixer has no values at this time
	return nil, false
}

/*
Helper function to determine if a channel has values before it has been closed.

It trys to read the first value of the channel and if it is sucessful copys the
entire channel into a new chan it returns.
*/
func hasValues(input chan MixerValue) (bool, chan MixerValue) {
	output := make(chan MixerValue)

	val, ok := <-input
	if ok {

		go func() {
			output <- val
			for val = range input {
				output <- val
			}
			close(output)
		}()

		return true, output
	} else {
		return false, input
	}
}

/*
Adds a channel to the list
*/
func addLink(in chan MixerValue, tail *chainLink) *chainLink {

	head := newChainLink()

	go func() {

		// The default value is nil for simplicity. It is never used.
		val := MixerValue(nil)
		// The available flag starts false so the link does not send the default value
		available := false

		for {
			select {
			case newVal, inOk := <-in:
				if inOk {
					// Store values that arive from input
					// Once there is data from the channel available is set to true and
					// and the link will put data on requests
					available = true
					val = newVal
				} else {
					// If the input channel is closed the link should "delete" itself and
					// exit the goroutine rather than continuing to respond on data
					head.replace <- tail
					return
				}
			case out := <-head.data:
				// Respond to requests for the current value
				if available {
					// Avoid sending the default value
					out <- val
				}
				if tail.data != nil {
					// If there are other links in the chain pass the request on
					tail.data <- out
				} else {
					// If this is the end of the chain close the channel to signal
					// it to the selector function
					close(out)
				}
			case tail = <-tail.replace:
				// Replace the tail if the next link is deleted
			}
		}
	}()

	return head
}

/*
Example Latest Takes Priority selector

This function returns the first value in the channel which is the value from
the most recent channel to be added and emit a value
*/
func LTP(input chan MixerValue) MixerValue {

	val := <-input

	for _ = range input {
	}

	return val
}
