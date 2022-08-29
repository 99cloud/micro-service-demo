package main

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/sirupsen/logrus"
)

var (
	//go:embed FangZhengFangSongJianTi-1.ttf
	fontData []byte
	fFont    *truetype.Font
)

func init() {
	font, err := freetype.ParseFont(fontData)
	if err != nil {
		logrus.Fatalf("freetype.ParseFont:%s", err)
	}
	fFont = font
}

func DrawText(img draw.Image, text string, fontSize float64, point image.Point) error {
	fctx := freetype.NewContext()
	fctx.SetFont(fFont)
	fctx.SetFontSize(fontSize)
	fctx.SetDst(img)
	fctx.SetDPI(200)
	fctx.SetClip(img.Bounds())
	fctx.SetSrc(&image.Uniform{C: color.RGBA{A: 255, R: 255, G: 255, B: 255}})
	_, err := fctx.DrawString(text, freetype.Pt(point.X, point.Y))
	if err != nil {
		return err
	}
	return nil
}
