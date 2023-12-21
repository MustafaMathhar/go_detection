package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
	//"github.com/vladimirvivien/go4vl/v4l2"
	//gocv "gocv.io/x/gocv"
)

func detectText(
	ctx context.Context,
	visionClient *vision.ImageAnnotatorClient,
	image *visionpb.Image,
	imageContext *visionpb.ImageContext,
	resultChan chan<- *visionpb.EntityAnnotation,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	res, err := visionClient.DetectTexts(ctx, image, imageContext, 1)
	if err != nil {
		log.Fatalf("Error sending requests: %v", err)
		return
	}

	if len(res) > 0 {

		if text := res[0].GetDescription(); len(text) > 0 {

			resultChan <- res[0] // Send the detected text annotation through the channel
		}
	}

}

func displayResults(
	res string,
	tc *texttospeech.Client,
	ctx context.Context,
	oc *oto.Context,
) {

	// Draw bounding boxes if needed
	// gocv.Rectangle(...)
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: res},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_FEMALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  0.5,
		},
	}
	time.Sleep(500 * time.Millisecond) // Adjust the duration as neededS

	resp, err := tc.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}

	fbReader := bytes.NewReader(resp.GetAudioContent())
	decodedMp3, err := mp3.NewDecoder(fbReader)
	player := oc.NewPlayer(decodedMp3)
	player.Play()

	// We can wait for the sound to finish playing using something like this
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}

}

func main() {

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

	// capture frame

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan

	client, err := texttospeech.NewClient(
		ctx,
		CREDENTIALS,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	const FRAMES_CAP = 30
	//var frameCount = 0
	var wg sync.WaitGroup

	visionClient, err := vision.NewImageAnnotatorClient(ctx, CREDENTIALS)
	if err != nil {
		log.Fatalf("Failed to create vision client: %v", err)
	}
	defer visionClient.Close()

	resultChan := make(chan *visionpb.EntityAnnotation)
	//var wg sync.WaitGroup
	var frameCount = 0
	go func() {
		defer close(resultChan)
		for frame := range dev.GetOutput() {
			frameCount = (frameCount + 1) % FRAMES_CAP
			if frameCount != 0 {
				continue
			}
			log.Printf("Frame is:  %d \n", len(frame))
			if err != nil {
				log.Println("test")

			}
			image := visionpb.Image{
				Content: frame,
			}

			wg.Add(1)
			go detectText(
				ctx,
				visionClient,
				&image,
				&imageContext,
				resultChan,

				&wg,
			) // Process detecte

		} /*
			for {
				/*if ok := webcam.Read(&img); !ok || img.Empty() {
					continue
				}
				frameCount = (frameCount + 1) % FRAMES_CAP
				if frameCount != 0 {
					continue
				}
				//go detectText(ctx, visionClient, img, &imageContext, resultChan)

				buf, err := gocv.IMEncode(gocv.JPEGFileExt, img)
				if err != nil {
					log.Fatalf("Error encoding frame to JPEG: %v", err)
				}
				imageData := buf.GetBytes()
				log.Printf("Image data size: %d", len(imageData))

				image := visionpb.Image{
					Content: imageData,
				}

				wg.Add(1)
				go detectText(
					ctx,
					visionClient,
					&image,
					&imageContext,
					resultChan,

					&wg,
				) // Process detected text annotation (change as needed)
				// Draw bounding boxes if needed
				// gocv.Rectangle(...)

				/*	window.IMShow(img)
					if window.WaitKey(1) >= 0 {
						break
					}*/

	}()

	for res := range resultChan {
		fmt.Print(res.GetDescription()) // Process detected text annotation (change as needed)
		// Draw bounding boxes if needed
		displayResults(res.GetDescription(), client, ctx, otoCtx)
		/*	window.IMShow(img)
			if window.WaitKey(1) >= 0 {
				break
			}*/
	}

}
