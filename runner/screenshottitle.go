package runner

import (
	"bytes"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
)

type Title struct {
	SuiteName  string
	CaseName   string
	Error      string
	ActionName string
	Step       int
}

func addTitle(screenshotFilename, targetFilename string, title Title) error {
	b, err := ioutil.ReadFile(screenshotFilename)
	if err != nil {
		return err
	}

	img, err := png.Decode(bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	const titleHeight = 80

	r := img.Bounds()
	r.Min.Y = titleHeight
	r.Max.Y += titleHeight + titleHeight
	newImg := image.NewRGBA(r)

	titleBackground := image.Rect(0, titleHeight, r.Max.X, titleHeight+titleHeight)
	black := color.RGBA{R: 12, A: 255}

	draw.Draw(newImg, titleBackground, &image.Uniform{C: black}, image.Point{}, draw.Src)

	var (
		o  = 10
		y  = 100
		ws = 10
		w  = 0
	)

	w += addBoldText(newImg, o+w, y, "Suite:")
	addText(newImg, o+w+ws, y, title.SuiteName)

	o, w, y = 10, 0, y+24
	w += addBoldText(newImg, o+w, y, "Case:")
	addText(newImg, o+w+ws, y, title.CaseName)

	o, w, y = 10, 0, y+24
	w += addBoldText(newImg, o+w, y, "Error:")
	addText(newImg, o+w+ws, y, title.Error)

	draw.Draw(newImg, r, img, image.Point{X: 0, Y: -titleHeight}, draw.Src)

	f, err := os.OpenFile(targetFilename, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}()

	return png.Encode(f, newImg)
}

func addText(img draw.Image, x, y int, text string) int { //nolint:unparam
	col := color.RGBA{R: 200, G: 100, A: 255}
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Regular8x16,
		Dot:  point,
	}
	d.DrawString(text)

	return d.MeasureString(text).Ceil()
}

func addBoldText(img draw.Image, x, y int, text string) int {
	col := color.RGBA{R: 200, G: 100, A: 255}
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Bold8x16,
		Dot:  point,
	}
	d.DrawString(text)

	return d.MeasureString(text).Ceil()
}
