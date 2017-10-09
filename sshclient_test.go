package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

//Changes it to actual password before running
var password = ""
var password2 = "back23433"

func TestNewClient(t *testing.T) {
	c, err := NewSSHClient("10.0.254.151", "back", password)
	checkError(err)
	defer c.Close()

	checkError(c.Expect(">", 5*time.Second))
	checkError(c.Send("en 5"))

	checkError(c.Expect("Password:", 5*time.Second))
	checkError(c.Send(password2))

	checkError(c.Expect("#", 5*time.Second))
	checkError(c.Send("copy run scp"))

	checkError(c.Expect("Destination filename [scp]?", 5*time.Second))
	checkError(c.Send("cloud@10.99.70.34"))

	checkError(c.Expect("over write?", 5*time.Second))
	checkError(c.Send(""))
	checkError(c.Send("exit"))

	//Expect AND, OR? How to do

}

func checkError(err error) {
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
