package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	vision "cloud.google.com/go/vision/apiv1"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  128000,
	WriteBufferSize: 128000,
}

func handleVideoUpload(
	w http.ResponseWriter,
	r *http.Request,
	vc *vision.ImageAnnotatorClient,
	tc *texttospeech.Client,
	ctx context.Context,
) {
	// Upgrade HTTP connection to a WebSocket connection

	// var wg sync.WaitGroup
	// resultChan := make(chan *visionpb.EntityAnnotation, 1024)
	readChan := make(chan []byte, 1)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	var prevText int

	go func() {
		for {
			select {
			default:
			case image, isOpen := <-readChan:
				if !isOpen {
					return
				}
				log.Println(len(image))
				img := buildImage(image)
				res, err := vc.DetectTexts(ctx, &img, &imageContext, 1)
				if err != nil {
					log.Fatalf("Error sending requests: %v", err)
					return
				}

				if len(res) > 0 {
					if text := res[0].GetDescription(); len(text) > 0 {
						if len(text) == prevText{
							continue
						}
						speechBytes := displayResults(text, tc, ctx)
						log.Println(
							len(speechBytes),
							text,
						)
						prevText = len(text)
						// Assuming 'conn' is defined somewhere in your code
						err := conn.WriteMessage(websocket.BinaryMessage, speechBytes)
						if err != nil {
							log.Println("Error sending message:", err)
							break
						}
					}
				}
			}
		}
	}()
	for {
		// Read message from the WebSocket client
		messageType, img, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		if messageType == websocket.BinaryMessage {
			readChan <- img
		}
		// Handle binary message (video data) received from the client
		//fmt.Printf("Received %d bytes of video data from client\n", len(img))
		//
		// wg.Add(1)
		// go detectText(ctx, vc, img, resultChan, &wg)
		// select {
		// case res := <-resultChan:
		// 	speechBytes := displayResults(res.GetDescription(), tc, ctx)
		// 	log.Println(
		// 		len(speechBytes),
		// 		res.GetDescription(),
		// 	)
		// 	err := conn.WriteMessage(websocket.BinaryMessage, speechBytes)
		// 	if err != nil {
		// 		log.Println("Error sending message:", err)
		// 		break
		// 	}
		// default:
		// 	// Drop the message if the channel is full (non-blocking)
		// }
		// Here, you can process or store the video data as needed
		// For demonstration purposes, we are just logging the size of received data
	}
}

func main() {
	ctx := context.Background()

	visionClient, err := vision.NewImageAnnotatorClient(ctx, CREDENTIALS)
	if err != nil {
		log.Fatalf("Failed to create vision client: %v", err)
	}
	defer visionClient.Close()

	tc, err := texttospeech.NewClient(
		ctx,
		CREDENTIALS,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer tc.Close()

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handleVideoUpload(w, r, visionClient, tc, ctx)

	})

	fmt.Println("WebSocket server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
