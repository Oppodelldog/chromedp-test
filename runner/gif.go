package runner

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	chromedptest "github.com/Oppodelldog/chromedp-test"
)

func createGIF(suiteID int, suiteName, testName string, screenshotOptions ScreenshotOptions) {
	var (
		name           = fmt.Sprintf("%v__", suiteID)
		imageFilenames = getImageFilenames(screenshotOptions.OutDir, name)
		origImages     = loadImages(imageFilenames)
		images         = make([]*image.Paletted, 0, len(origImages))
		delay          = make([]int, 0, len(origImages))
	)

	if len(origImages) == 0 {
		chromedptest.Printf("Did not find any png files for %s\n", name)

		return
	}

	for _, origImg := range origImages {
		palettedImage := image.NewPaletted(origImg.Bounds(), palette.Plan9)
		draw.Draw(palettedImage, palettedImage.Rect, origImg, origImg.Bounds().Min, draw.Over)
		images = append(images, palettedImage)
		delay = append(delay, 100)
	}

	anim := gif.GIF{Delay: delay, Image: images}

	gifFilename := fmt.Sprintf("%s%s-%s.gif", name, suiteName, testName)
	gifFilePath := path.Join(screenshotOptions.PostProcessing.OutDir, gifFilename)

	f, err := os.OpenFile(gifFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = gif.EncodeAll(f, &anim)
	if err != nil {
		panic(err)
	}

	if screenshotOptions.PostProcessing.RemoveImages {
		for _, imageFilename := range imageFilenames {
			err := os.Remove(imageFilename)
			if err != nil {
				panic(err)
			}
		}
	}
}

func loadImages(filenames []string) []image.Image {
	var images = make([]image.Image, 0, len(filenames))

	for _, filename := range filenames {
		f, err := os.OpenFile(filename, syscall.O_RDONLY, 0600)
		if err != nil {
			panic(err)
		}

		img, err := png.Decode(f)
		if err != nil {
			panic(err)
		}

		err = f.Close()
		if err != nil {
			panic(err)
		}

		images = append(images, img)
	}

	return images
}

func getImageFilenames(dir, filenamePrefix string) []string {
	var imgFilenames []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(info.Name(), filenamePrefix) {
			if filepath.Ext(path) == ".png" {
				imgFilenames = append(imgFilenames, path)
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return imgFilenames
}
