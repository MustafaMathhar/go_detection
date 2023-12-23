package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"

	"github.com/gorilla/websocket"
)

const (
	serverURL      = "ws://localhost:8080/upload" // Change to your server's WebSocket URL
	frameWidth     = 640
	frameHeight    = 480
	frameChannels  = 3
	waitKeyDelayMS = 1000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  128000,
	WriteBufferSize: 128000,
}

func main() {
	// Open the webcam device
	ctx := context.Background()
	dev, err := device.Open("/dev/video0", device.WithPixFormat(
		v4l2.PixFormat{PixelFormat: v4l2.PixelFmtJPEG, Width: 480, Height: 360},
	))
	if err != nil {
		log.Fatalf("the error: %d ", err)
	}
	defer dev.Close()
	if err := dev.Start(ctx); err != nil {
		log.Fatalf("failed to start stream: %s", err)
	}

	u, _ := url.Parse(serverURL)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Initialize image matrix to store webcam frames

	// Create a signal channel to handle interruption (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	messageChan := make(chan []byte,128000) // Channel to handle received messages

	go func() {
		for {
			messageType, sound, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				close(messageChan)
				return
			}

			if messageType == websocket.BinaryMessage && len(sound) > 0 {
				log.Println("sound bytes are:", len(sound))
				// Handle binary message (sound data) received from the server
				// Process or use the received sound data as needed
				select {
				case messageChan <- sound:
				default:
					// Drop the message if the channel is full (non-blocking)
					log.Println("Message dropped: channel full")
				}
			}
		}
	}()
  for {
		select {
		case frame, ok := <-dev.GetOutput():
			if !ok {
				fmt.Println("Frame channel closed")
				return
			}
			log.Printf("Frame is:  %d \n", len(frame))
			if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
				log.Println("Error sending message:", err)
				return
			}
		case <-interrupt:
			fmt.Println("Streaming stopped")
			return
		case sound, ok := <-messageChan:
			if !ok {
				fmt.Println("Message channel closed")
				return
			}
			log.Println("Received sound bytes:", len(sound))
			// Process or use the received sound data from the server
			// ...
		}
		time.Sleep(time.Millisecond * waitKeyDelayMS)
	}
	
}
