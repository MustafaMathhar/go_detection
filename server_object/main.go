package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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
	readChan := make(chan []byte, 4096)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	var prevText int

	go func() {
		for v := range readChan {
			img := buildImage(v)

			finalTxt, _:= detectObjects(ctx, &img, &imageContext, vc)
			//finalTxt, locale := detectText(&img, &imageContext, vc, ctx)
			if len(finalTxt) == prevText {
				continue
			}
      

			speechBytes := displayResults(finalTxt,"", tc, ctx)
			log.Println(
				len(speechBytes),
				finalTxt,
			)
			prevText = len(finalTxt)
			// Assuming 'conn' is defined somewhere in your code
			err = conn.WriteMessage(websocket.BinaryMessage, speechBytes)
			if err != nil {
				log.Println("Error sending message:", err)
				break
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
	}
}


func main() {
   value, ok := os.LookupEnv("GCLOUD_CREDENTIALS") 

  if !ok{
    log.Fatal("Error: no credentials files were found, please set it using \"GCLOUD_CREDENTIALS\" variable.")
  }
  wordPtr:=flag.String("ip", "localhost", "the ip address")
  flag.Parse()
  if wordPtr==nil{
    log.Fatal("Error no address set")
  }
	ctx := context.Background()

	visionClient, err := vision.NewImageAnnotatorClient(ctx, InitializeCredentials(value))
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
  
	serverAddr := *wordPtr + ":8080"
	fmt.Println("WebSocket server is running on port 8080")
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
