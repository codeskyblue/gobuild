package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"net/http"

	"code.google.com/p/freetype-go/freetype"
)

const (
	fontFile = "bin/font.ttf"
	fontSize = 12
	fontDPI  = 72
)

var (
	black color.Color = color.RGBA{0, 0, 0, 255}
	green color.Color = color.RGBA{0, 255, 0, 255}
)

func initBadge() {
	m.Get("/badge/:web/:name/:p/download.png", func(w http.ResponseWriter, r *http.Request) {
		var err error
		img := image.NewNRGBA(image.Rect(0, 0, 180, 16))
		fontBytes, err := ioutil.ReadFile(fontFile)
		if err != nil {
			return
		}
		font, err := freetype.ParseFont(fontBytes)
		if err != nil {
			return
		}
		left, right := img.Bounds(), img.Bounds()
		const middle = 65
		left.Max = image.Pt(middle, 16)
		right.Min = image.Pt(middle, 0)
		// fill left(black) right(green)
		draw.Draw(img, left, &image.Uniform{black}, image.ZP, draw.Src)
		draw.Draw(img, right, &image.Uniform{green}, image.ZP, draw.Src)

		// draw "gobuild.io | download"
		c := freetype.NewContext()
		c.SetDPI(fontDPI)
		c.SetFont(font)
		c.SetFontSize(fontSize)
		c.SetClip(img.Bounds())
		c.SetDst(img)
		c.SetSrc(image.White)
		pt := freetype.Pt(5, 12)
		_, err = c.DrawString("gobuild.io", pt) // 10 chars width = 60px
		if err != nil {
			return
		}
		c.SetSrc(image.Black)
		pt = freetype.Pt(middle+20, 12)
		_, err = c.DrawString("download", pt)

		w.Header().Set("Content-Type", "image/png")
		png.Encode(w, img)
	})
}
