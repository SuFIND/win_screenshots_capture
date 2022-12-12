package util

import (
	"errors"
	"image"
	"image/jpeg"
	"os"
)

func CreateImage(rect image.Rectangle) (img *image.RGBA, e error) {
	img = nil
	e = errors.New("Cannot create image.RGBA")

	defer func() {
		err := recover()
		if err == nil {
			e = nil
		}
	}()
	// image.NewRGBA may panic if rect is too large.
	img = image.NewRGBA(rect)

	return img, e
}

func SaveJPEG(img *image.RGBA, filePath string, imgQuality int) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	err = jpeg.Encode(file, img, &jpeg.Options{Quality: imgQuality})
	if err != nil {
		return
	}
}
