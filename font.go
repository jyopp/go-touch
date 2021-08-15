package main

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

const (
	systemFont     = "goregular"
	systemBoldFont = "gobold"
)

var (
	// These global values affect how fonts are rendered.
	// These values should be mutated only before fonts are loaded.
	fontDPI                  float64 = 96
	fontSubpixelQuantization         = 2
)

type fontKey struct {
	name      string
	size, dpi float64
}

var (
	_ttfCache     map[string]*truetype.Font
	_ttfProviders map[string]func() []byte
	_fontCache    map[fontKey]*Font
	_cacheMutex   sync.Mutex
)

func registerTTF(name string, provider func() []byte) {
	_ttfProviders[name] = provider
}

func loadTTF(name string) *truetype.Font {
	if font, ok := _ttfCache[name]; ok {
		return font
	}

	if provider, ok := _ttfProviders[name]; ok {
		if font, err := truetype.Parse(provider()); err == nil {
			_ttfCache[name] = font
			return font
		}
	}
	return nil
}

func init() {
	_ttfCache = make(map[string]*truetype.Font)
	_ttfProviders = map[string]func() []byte{}
	_fontCache = make(map[fontKey]*Font)

	registerTTF(systemFont, func() []byte {
		return goregular.TTF
	})
	registerTTF(systemBoldFont, func() []byte {
		return gobold.TTF
	})
}

type Font struct {
	Face    font.Face
	Metrics font.Metrics
	m       sync.Mutex
	ctx     *freetype.Context
}

func SharedFont(name string, size float64) *Font {
	_cacheMutex.Lock()
	defer _cacheMutex.Unlock()

	cacheKey := fontKey{name, size, fontDPI}
	if f, ok := _fontCache[cacheKey]; ok {
		return f
	}
	// TODO: This could be more threadsafe with a sync.Map
	font := &Font{}
	font.Init(name, size)
	if font.Face == nil {
		return nil
	}

	_fontCache[cacheKey] = font
	return font
}

func (f *Font) Init(name string, size float64) {
	f.m.Lock()
	defer f.m.Unlock()

	opts := truetype.Options{
		Size:       size,
		DPI:        fontDPI,
		SubPixelsX: fontSubpixelQuantization,
		SubPixelsY: fontSubpixelQuantization,
	}

	font := loadTTF(name)
	if font == nil {
		return
	}
	if f.Face = truetype.NewFace(font, &opts); f.Face == nil {
		return
	}
	f.Metrics = f.Face.Metrics()

	ctx := freetype.NewContext()
	ctx.SetDPI(opts.DPI)
	ctx.SetHinting(opts.Hinting)
	ctx.SetFontSize(opts.Size)
	// Set font last to avoid thrashing ctx's internal cache
	ctx.SetFont(font)
	f.ctx = ctx
}

func (f *Font) Measure(text string, maxsize image.Point) image.Point {
	f.m.Lock()
	defer f.m.Unlock()

	size := image.Point{
		X: font.MeasureString(f.Face, text).Ceil(),
		Y: (f.Metrics.Ascent + f.Metrics.Descent).Ceil(),
	}
	if size.X > maxsize.X {
		size.X = maxsize.X
	}
	if size.Y > maxsize.Y {
		size.Y = maxsize.Y
	}
	// println("Measured", text, "in", size.String())
	return size
}

func (f *Font) Draw(img draw.Image, text string, rect image.Rectangle, c color.Color) error {
	f.m.Lock()
	defer f.m.Unlock()

	f.ctx.SetSrc(image.NewUniform(c))
	f.ctx.SetDst(img)
	f.ctx.SetClip(rect)

	textOrigin := fixed.Point26_6{
		X: fixed.I(rect.Min.X),
		Y: fixed.I(rect.Min.Y) + f.Metrics.Ascent,
	}
	_, err := f.ctx.DrawString(text, textOrigin)
	// println("Rendered", text, "in", rect.String(), err)
	return err
}
