package main

import (
	"context"
	"log"
	"sync"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"google.golang.org/api/option"
)

var CREDENTIALS = option.WithCredentialsFile("./glassy-courage-399211-c6db5fa7335d.json")
var imageContext = visionpb.ImageContext{
	LanguageHints: []string{"en", "ar"},
}

func buildImage(image []byte) visionpb.Image {
	return visionpb.Image{
		Content: image,
	}

}
func detectText(

	ctx context.Context,
	visionClient *vision.ImageAnnotatorClient,
	image []byte,
	resultChan chan<- *visionpb.EntityAnnotation,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	img := buildImage(image)
	res, err := visionClient.DetectTexts(ctx, &img, &imageContext, 1)
	if err != nil {
		log.Fatalf("Error sending requests: %v", err)
		return
	}

	if len(res) > 0 {

		if text := res[0].GetDescription(); len(text) > 0 {
			//log.Println(text)

			resultChan <- res[0] // Send the detected text annotation through the channel
		}
	}

}

func createTTSRequest(res string) texttospeechpb.SynthesizeSpeechRequest {

	return texttospeechpb.SynthesizeSpeechRequest{
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

}

func displayResults(
	res string,
	tc *texttospeech.Client,
	ctx context.Context,
	//oc *oto.Context,
) []byte {

	// Draw bounding boxes if needed
	// gocv.Rectangle(...)
	req := createTTSRequest(res)
	resp, err := tc.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Fatal(err)
	}
	return resp.GetAudioContent()

	/*c	fbReader := bytes.NewReader(resp.GetAudioContent())
	decodedMp3, err := mp3.NewDecoder(fbReader)
	player := oc.NewPlayer(decodedMp3)
	player.Play()

	// We can wait for the sound to finish playing using something like this
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}*/

}
