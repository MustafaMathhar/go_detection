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

var CREDENTIALS = option.WithCredentialsFile("./coastal-mercury-410017-dca4a85de2f2.json")
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
) (string, string) {

	res, err := vc.DetectTexts(ctx, img, imageContext, 1)

	if err != nil {
		log.Fatalf("Error sending requests: %v", err)
	}

	if len(res) <= 0 {
		return "", ""
	}
	text, locale := res[0].GetDescription(), res[0].GetLocale()
	if len(text) <= 0 {
		//log.Println(text)
		return "", ""
	}
	return text, locale
}

func createTTSRequest(res string,locale string) texttospeechpb.SynthesizeSpeechRequest {
  var  langCode string
  if locale=="ar" {
    langCode="ar-XA"
    
  }else{
    langCode="en-US"
  }

	return texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: res},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name:         langCode+"-Wavenet-A",
			LanguageCode: langCode,
			//SsmlGender:   texttospeechpb.,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:    texttospeechpb.AudioEncoding_MP3,
			EffectsProfileId: []string{"headphone-class-device"},
			SpeakingRate:     1.5,
			Pitch:            1,
		},
	}

}

func displayResults(
	res string,
  locale string,
	tc *texttospeech.Client,
	ctx context.Context,
	//oc *oto.Context,
) []byte {

	// Draw bounding boxes if needed
	// gocv.Rectangle(...)
	req := createTTSRequest(res,locale)
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
