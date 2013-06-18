/*
Broadcasts a DMX Value to one or more DMX Channels. Needs rewriting to use
closed channels instead of timeouts.
*/
package dmx

import "time"

const (
  defaultValue DMXValue = 0
  timeoutLimit = 1 * time.Second
)

type DMXParam struct {
  label string
  value DMXValue
  subscribers map[chan DMXValue] bool
}

func NewDMXParam(label string) *DMXParam {
  param := new(DMXParam)

  param.value = defaultValue
  param.label = label

  param.subscribers = make(map[chan DMXValue] bool)

  return param
}

func (param *DMXParam) Label() string {
  return param.label
}

func (param *DMXParam) Value() DMXValue {
  return param.value
}

func (param *DMXParam) SetValue(val DMXValue) {
  param.value = val

  for sub, _ := range param.subscribers {
    timeout := time.After(timeoutLimit)

    go func() {

      select {
      case sub <- val:
      case _ = <-timeout:
        delete(param.subscribers, sub)
      }
    }()
  }
}

func (param *DMXParam) Subscribe() chan DMXValue {
  c := make(chan DMXValue)
  param.subscribers[c] = true
  return c
}
