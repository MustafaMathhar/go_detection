package main

import (
	"bytes"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

func showImageInWindow(frameBytes []byte) error {
	img, _, err := image.Decode(bytes.NewReader(frameBytes))
	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}

	ebiten.SetWindowSize(img.Bounds().Max.X, img.Bounds().Max.Y)
	ebiten.SetWindowTitle("Image Window")

	update := func(screen *ebiten.Image) error {
    imger := ebiten.NewImageFromImage(img)
		screen.DrawImage(imger, nil)
		return nil
	}

	if err := ebiten.RunGame(&game{update: update}); err != nil {
		return fmt.Errorf("error running the game: %v", err)
	}
	return nil
}

type game struct {
	update func(screen *ebiten.Image) error
}

func (g *game) Update() error {
	return g.update(nil)
}

func (g *game) Draw(screen *ebiten.Image) {
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
