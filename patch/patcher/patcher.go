package main

import (
  "golx/patch"
  "fmt"
)

type forwarder struct {
  input chan string
  output chan string
}

func newForwarder() *forwarder {
  f := new(forwarder)
  f.input = make(chan string)
  f.output = make(chan string)

  go func() {
    for v := range(f.input) {
      f.output <- v
    }
  }()

  return f
}

func (f *forwarder) InputChannel() chan string {
  return f.input
}

func (f *forwarder) OutputChannel() chan string {
  return f.output
}

func main() {
  a := newForwarder()
  b := newForwarder()
  c := newForwarder()
  patch.Patch(a, b)

  go func() { a.input <- "Hello World" }()

  v := <-b.output

  fmt.Println(v)

  patch.Patch(b, c)

  go func() { a.input <- "Hello Again" }()

  v = <-c.output
  fmt.Println(v)

  patch.Unpatch(a, b)
  patch.Unpatch(b, c)

  patch.Patch(c, b)

  go func() { c.input <- "Output" }()

  v = <-b.output
  fmt.Println(v)

}
