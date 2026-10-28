// Package raster renders PDF pages to images with pdfium compiled to
// WebAssembly, so the app needs no native PDF library.
package raster

import (
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/tetratelabs/wazero"
)

// Renderer owns a pdfium instance. It is safe for concurrent use; renders
// are serialized.
type Renderer struct {
	mu       sync.Mutex
	pool     pdfium.Pool
	instance pdfium.Pdfium
}

// New starts pdfium. Compiling the WebAssembly module takes about a second,
// so create one Renderer and reuse it.
func New() (*Renderer, error) {
	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:  1,
		MaxIdle:  1,
		MaxTotal: 1,
		// Documents are passed as bytes; pdfium needs no file system access.
		FSConfig: wazero.NewFSConfig(),
	})
	if err != nil {
		return nil, fmt.Errorf("start pdfium: %w", err)
	}
	instance, err := pool.GetInstance(30 * time.Second)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("start pdfium: %w", err)
	}
	return &Renderer{pool: pool, instance: instance}, nil
}

func (r *Renderer) Close() error {
	r.instance.Close()
	return r.pool.Close()
}

// Format selects the pixel format to render in.
type Format int

const (
	Color Format = iota
	Gray
)

// Render draws every page of pdf at dpi and calls fn with each image.
// The image is only valid during the call; copy or encode it there.
func (r *Renderer) Render(pdf []byte, dpi int, format Format, fn func(page int, img image.Image) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, err := r.instance.OpenDocument(&requests.OpenDocument{File: &pdf})
	if err != nil {
		return fmt.Errorf("open PDF for rendering: %w", err)
	}
	defer r.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	count, err := r.instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return err
	}

	imgFormat := requests.RenderImageFormatRGBA
	if format == Gray {
		imgFormat = requests.RenderImageFormatGrayscale
	}
	for i := range count.PageCount {
		res, err := r.instance.RenderPageInDPI(&requests.RenderPageInDPI{
			Page:        requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: i}},
			DPI:         dpi,
			ImageFormat: imgFormat,
		})
		if err != nil {
			return fmt.Errorf("render page %d: %w", i+1, err)
		}
		err = fn(i+1, res.Result.RenderedImage)
		res.Cleanup()
		if err != nil {
			return err
		}
	}
	return nil
}
