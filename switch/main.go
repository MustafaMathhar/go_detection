package main

import (
	"fmt"
	"os"
	"time"

	rpio "github.com/stianeikeland/go-rpio/v4"
)

func main() {
	// Open Raspberry Pi GPIO
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()

	// Use GPIO pin 17 (BCM numbering) as an input for a switch
	pin := rpio.Pin(17)
	pin.Input()

	// Loop to continuously read the state of the switch
	for {
		if pin.Read() == rpio.High {
			fmt.Println("Switch pressed")
		} else {
			fmt.Println("Switch not pressed")
		}

		time.Sleep(100 * time.Millisecond) // Add a short delay to avoid continuous polling
	}
}
