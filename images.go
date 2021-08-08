package main

import (
	"bytes"
	"embed"
	"image"

	_ "image/png"
)

//go:embed images/*

var Images embed.FS

func ReadImage(name string) (image.Image, error) {
	data, err := Images.ReadFile("images/" + name)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}
