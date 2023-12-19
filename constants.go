package main

import (
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/ebitengine/oto/v3"
	"google.golang.org/api/option"
)


var CREDENTIALS = option.WithCredentialsFile("./glassy-courage-399211-c6db5fa7335d.json")
var imageContext = visionpb.ImageContext{
	LanguageHints: []string{"en", "ar"},
}
var op = &oto.NewContextOptions{
	SampleRate:   44100,
	ChannelCount: 2,
	Format:       oto.FormatSignedInt16LE,
}

