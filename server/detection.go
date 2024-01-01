package main

import (
	"context"
	"errors"
	"log"

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

func detectObjects(
	ctx context.Context,
	img *visionpb.Image,
	imageContext *visionpb.ImageContext,
	vc *vision.ImageAnnotatorClient,
) (string, error) {
	res, err := vc.LocalizeObjects(ctx, img, imageContext)
	if err != nil {
		log.Fatalf("Error sending requests: %v", err)
	}
	if len(res) <= 0 {
		return "", errors.New("response is 0")

	}
	var finalTxt string
	for _, annotation := range res {
		if annotation.GetScore() >= 0.60 {

			finalTxt = finalTxt + annotation.GetName() + " "
		}
	}
	return finalTxt, nil
}
func detectText(

	img *visionpb.Image,
	imageContext *visionpb.ImageContext,
	vc *vision.ImageAnnotatorClient,
	ctx context.Context,
) string {

	res, err := vc.DetectTexts(ctx, img, imageContext, 1)
	if err != nil {
		log.Fatalf("Error sending requests: %v", err)
	}

	if len(res) <= 0 {
		return ""
	}
	text := res[0].GetDescription()
	if len(text) <= 0 {
		//log.Println(text)
		return ""
	}
	return text
}

func createTTSRequest(res string) texttospeechpb.SynthesizeSpeechRequest {

	return texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: res},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name:         "en-US-Wavenet-A",
			LanguageCode: "en-US",
			//SsmlGender:   texttospeechpb.,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:    texttospeechpb.AudioEncoding_MP3,
			EffectsProfileId: []string{"headphone-class-device"},
			SpeakingRate:     0.6,
			Pitch:            0.5,
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
