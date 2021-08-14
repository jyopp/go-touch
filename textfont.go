package main

import (
	"fmt"
	"image"
	"image/color"
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
	fontSubpixelQuantization         = 4
)

var _ttfCache map[string]*truetype.Font
var _ttfProviders map[string]func() []byte
var _fontCache map[string]*Font

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
	_fontCache = make(map[string]*Font)

	registerTTF(systemFont, func() []byte {
		return goregular.TTF
	})
	registerTTF(systemBoldFont, func() []byte {
		return gobold.TTF
	})
}

type Font struct {
	font *truetype.Font
	face font.Face
	m    sync.Mutex
	opts truetype.Options
}

func SharedFont(name string, size float64) *Font {
	cacheName := fmt.Sprintf("%s-%0.1f-%0f", name, size, fontDPI)
	if f, ok := _fontCache[cacheName]; ok {
		return f
	}
	// TODO: This could be more threadsafe with a sync.Map
	font := &Font{}
	font.Init(name, size)
	if font.face == nil {
		return nil
	}

	_fontCache[cacheName] = font
	return font
}

func (f *Font) Init(name string, size float64) {
	f.m.Lock()
	defer f.m.Unlock()

	f.opts.Size = size
	f.opts.DPI = fontDPI
	f.opts.Hinting = font.HintingFull
	f.opts.SubPixelsX = fontSubpixelQuantization
	f.opts.SubPixelsY = fontSubpixelQuantization

	if f.font = loadTTF(name); f.font != nil {
		f.face = truetype.NewFace(f.font, &f.opts)
	}

	// Release all resources if initialization fails
	if f.font == nil || f.face == nil {
		// Maybe panic instead?
		f.font, f.face = nil, nil
		f.opts = truetype.Options{}
	}
}

func (f *Font) Render(text string, size image.Point, alpha uint8) *image.Alpha {
	metrics := f.face.Metrics()
	advance := font.MeasureString(f.face, text)
	// Extent is in points and must be converted to pixels.
	renderBounds := image.Rectangle{
		Max: image.Point{
			X: advance.Ceil(),
			// Use one line for now; font.BoundString is not consistent.
			Y: (metrics.Ascent + metrics.Descent).Ceil(),
		},
	}
	if renderBounds.Max.X > size.X {
		renderBounds.Max.X = size.X
	}
	if renderBounds.Max.Y > size.Y {
		renderBounds.Max.Y = size.Y
	}

	img := image.NewAlpha(renderBounds)

	ctx := freetype.NewContext()
	ctx.SetSrc(image.NewUniform(color.Alpha{alpha}))
	ctx.SetDst(img)
	ctx.SetClip(renderBounds)
	ctx.SetDPI(f.opts.DPI)
	ctx.SetHinting(f.opts.Hinting)
	ctx.SetFontSize(f.opts.Size)
	// Set font last to avoid thrashing ctx's internal cache
	ctx.SetFont(f.font)

	// BoundString isn't doing multiline drawing
	textOrigin := fixed.Point26_6{
		X: 0,
		Y: metrics.Ascent,
	}
	if _, err := ctx.DrawString(text, textOrigin); err == nil {
		// println("Rendered", text, "in", size.String(), ":", img.Rect.String())
		return img
	} else {
		println("Failed to render", text, ":", err)
	}
	return nil
}

// RenderedText caches the alphamask and dimensions of a text string
type RenderedText struct {
	*image.Alpha
	font    *Font
	MaxSize image.Point
	Text    string
}

func (rt *RenderedText) Invalidate() {
	rt.Alpha = nil
}

func (rt *RenderedText) SetFont(name string, size float64) {
	f := SharedFont(name, size)
	if f == rt.font {
		return
	}
	rt.font = f
	rt.Alpha = nil
}

func (rt *RenderedText) Render() {
	rt.Alpha = rt.font.Render(rt.Text, rt.MaxSize, 0xFF)
}

func (rt *RenderedText) Prepare(text string, maxSize image.Point) *image.Alpha {
	if rt.Alpha == nil || !maxSize.Eq(rt.MaxSize) || text != rt.Text {
		rt.MaxSize = maxSize
		rt.Text = text
		rt.Render()
	}
	return rt.Alpha
}
