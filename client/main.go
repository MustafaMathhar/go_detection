package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

const (
	frameWidth     = 480
	frameHeight    = 360
	waitKeyDelayMS = 300
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  128000,
	WriteBufferSize: 128000,
}

func main() {
	// open sound

	// Initialize PortAudio

	serverIP := flag.String("ip", "", "ip of the server")
	flag.Parse()
	if len(*serverIP) == 0 {
		log.Fatal("Server has no ip address specified")
	}
	// Open the webcam device
	ctx := context.Background()
	dev, err := device.Open("/dev/video0", device.WithPixFormat(
		v4l2.PixFormat{PixelFormat: v4l2.PixelFmtJPEG, Width: frameWidth, Height: frameHeight},
	))
	if err != nil {
		log.Fatalf("the error: %d ", err)
	}
	defer dev.Close()
	if err := dev.Start(ctx); err != nil {
		log.Fatalf("failed to start stream: %s", err)
	}
	serverAddr := fmt.Sprintf("ws://%s:8080/upload", *serverIP)
	u, _ := url.Parse(serverAddr)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Initialize image matrix to store webcam frames

	// Create a signal channel to handle interruption (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	messageChan := make(chan []byte,  1) // Channel to handle received messages

	go func() {
		for {
			messageType, sound, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				close(messageChan)
				return
			}

			if messageType == websocket.BinaryMessage && len(sound) > 0 {
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
			dev.Close()
			fmt.Println("Streaming stopped")
			return
		case sound, ok := <-messageChan:
			if !ok {
				fmt.Println("Message channel closed")
				return
			}
			log.Println("Received sound bytes:", len(sound))
			tempFile, err := os.CreateTemp("", "sound.wav")
			if err != nil {
				log.Fatal("failed to create temp file")

			}
			defer os.Remove(
				tempFile.Name(),
			) // Remove the temporary file when done		log.Fatalf("PortAudio stream error: %v", err)
			_, err = tempFile.Write(sound)

			if err != nil {
				log.Fatal("Error writing to file:", err)
			}

			if err != nil {
				log.Fatal("Error writing to file:", err)
			}
			command := "/usr/bin/paplay"
			args := []string{
				"paplay",
				"--channels=2",
				"--rate=44100",
				"--format=s16le",
				tempFile.Name(),
			}
			pid, err := syscall.ForkExec(command, args, &syscall.ProcAttr{
				Env:   os.Environ(),
				Files: []uintptr{0, 1, 2}, // Inherit stdin, stdout, stderr
			})
			if err != nil {
				log.Fatalf("Error executing paplay: %s", err)
			}
			var ws syscall.WaitStatus
			syscall.Wait4(pid, &ws, 0, nil)
			log.Printf("Process exited with status %d\n", ws.ExitStatus())
		}
		time.Sleep(time.Millisecond * waitKeyDelayMS)
	}
}
