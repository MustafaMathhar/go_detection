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
	pin := rpio.Pin(22)
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge)
	// Loop to continuously read the state of the switch
	for {
		if pin.EdgeDetected() {
			fmt.Println("Switch pressed")
			pin.Detect(rpio.NoEdge)            // Disable further edge detection for now
			time.Sleep(100 * time.Millisecond) // Sleep to debounce the switch
			for pin.Read() == rpio.Low {
				// Wait for the switch to be released
				time.Sleep(100 * time.Millisecond)
			}
			pin.Detect(rpio.FallEdge) // Re-enable edge detection
		}

		time.Sleep(100 * time.Millisecond) // Add a short delay to avoid continuous polling
	}
}
