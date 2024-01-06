package main

import (
	"fmt"
	"log"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

// Open Raspberry Pi GPIO

func main() {
	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Lookup a pin by its number:
	p := gpioreg.ByName("GPIO22")
	if p == nil {
		log.Fatal("Failed to find GPIO2")
	}

	fmt.Printf("%s: %s\n", p, p.Function())

	// Set it as input, with an internal pull down resistor:
	if err := p.In(gpio.PullDown, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}

	// Wait for edges as detected by the hardware, and print the value read:
	for {
		p.WaitForEdge(-1)
		fmt.Printf("-> %s\n", p.Read())
	}
}
