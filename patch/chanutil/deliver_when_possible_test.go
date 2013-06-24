package chanutil

import "testing"

func TestDeliverWhenPossible(t *testing.T) {
  input := make(chan int)
  output := make(chan int)

  err := DeliverWhenPossible(input, output)

  if err != nil {
    t.Log("Error initializing DeliverWhenPossible: ", err.Error())
    t.FailNow()
  }

  // Output delivers values sent on input
  input <- 123
  val := <-output

  if val != 123 {
    t.Log("Did not recieve value sent on input")
    close(input)
    t.FailNow()
  }

  // Output can skip values sent on input if they are not read
  input <- 100
  input <- 200

  val = <-output

  if val != 200 {
    t.Log("Did not recieve most recent value sent on input")
    t.Fail()
  }

  close(input)
}
