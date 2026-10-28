package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"printer/internal/book"
	"printer/internal/impose"
	"printer/internal/job"
	"printer/internal/printing"
	"printer/internal/raster"
)

// App is the backend the UI talks to. Every exported method is callable
// from the frontend.
type App struct {
	ctx context.Context

	mu       sync.Mutex
	renderer *raster.Renderer
	doc      *job.Document
	cfg      Config
}

func NewApp() *App {
	return &App{cfg: loadConfig()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Warm up pdfium in the background so the first file opens quickly.
	go func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.rendererLocked()
	}()
}

func (a *App) shutdown(context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.renderer != nil {
		a.renderer.Close()
	}
}

func (a *App) rendererLocked() (*raster.Renderer, error) {
	if a.renderer == nil {
		r, err := raster.New()
		if err != nil {
			return nil, err
		}
		a.renderer = r
	}
	return a.renderer, nil
}

// Options is everything the UI needs to draw its controls.
type Options struct {
	Papers      []book.Paper       `json:"papers"`
	Config      Config             `json:"config"`
	Printers    []printing.Printer `json:"printers"`
	DirectPrint bool               `json:"directPrint"`
	Defaults    book.Settings      `json:"defaults"`
}

func (a *App) Options() Options {
	printers, _ := printing.Printers()
	a.mu.Lock()
	defer a.mu.Unlock()
	return Options{
		Papers:      book.Papers,
		Config:      a.cfg,
		Printers:    printers,
		DirectPrint: printing.DirectSupported(),
		Defaults:    book.DefaultSettings(),
	}
}

// StartupFile is the PDF passed on the command line ("Open with…"), or "".
func (a *App) StartupFile() string {
	if len(os.Args) > 1 && strings.EqualFold(filepath.Ext(os.Args[1]), ".pdf") {
		return os.Args[1]
	}
	return ""
}

// ChooseFile asks for a PDF and opens it. It returns nil when cancelled.
func (a *App) ChooseFile() (*job.Info, error) {
	a.mu.Lock()
	dir := ""
	if a.cfg.LastFile != "" {
		dir = filepath.Dir(a.cfg.LastFile)
	}
	a.mu.Unlock()
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open book PDF",
		DefaultDirectory: dir,
		Filters:          []runtime.FileFilter{{DisplayName: "PDF documents", Pattern: "*.pdf"}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	return a.OpenFile(path)
}

// OpenFile loads a PDF by path (used for drag and drop and reopening).
func (a *App) OpenFile(path string) (*job.Info, error) {
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		return nil, fmt.Errorf("%s is not a PDF", filepath.Base(path))
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	r, err := a.rendererLocked()
	if err != nil {
		return nil, err
	}
	doc, err := job.Load(a.ctx, r, path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %w", filepath.Base(path), err)
	}
	a.doc = doc
	a.cfg.LastFile = path
	saveConfig(a.cfg)
	info := doc.Info()
	return &info, nil
}

// Reload re-reads the current file after it was re-exported.
func (a *App) Reload() (*job.Info, error) {
	a.mu.Lock()
	doc := a.doc
	a.mu.Unlock()
	if doc == nil {
		return nil, errors.New("no file is open")
	}
	return a.OpenFile(doc.Path)
}

// Analyze checks the open document with the given settings and renders the
// sheet previews. The settings are remembered for next time.
func (a *App) Analyze(s book.Settings) (*job.Result, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.doc == nil {
		return nil, errors.New("no file is open")
	}
	r, err := a.rendererLocked()
	if err != nil {
		return nil, err
	}
	res, err := a.doc.Analyze(a.ctx, r, s)
	if err != nil {
		return nil, err
	}
	a.cfg.Settings = s
	saveConfig(a.cfg)
	return res, nil
}

// SavePDF writes the print-ready PDF where the user chooses. It returns the
// path, or "" when cancelled.
func (a *App) SavePDF(s book.Settings) (string, error) {
	a.mu.Lock()
	doc := a.doc
	a.mu.Unlock()
	if doc == nil {
		return "", errors.New("no file is open")
	}
	base := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Save print-ready PDF",
		DefaultDirectory: filepath.Dir(doc.Path),
		DefaultFilename:  base + " - print.pdf",
		Filters:          []runtime.FileFilter{{DisplayName: "PDF documents", Pattern: "*.pdf"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := a.writeFile(doc, s, impose.Options{Sides: impose.AllSides}, path); err != nil {
		return "", err
	}
	return path, nil
}

// PrintRequest describes one trip to the printer.
type PrintRequest struct {
	Settings book.Settings `json:"settings"`
	Printer  string        `json:"printer"`
	// Pass is "all" for automatic duplex, or "fronts" / "backs" for the two
	// passes of manual duplex.
	Pass         impose.Sides `json:"pass"`
	ReverseBacks bool         `json:"reverseBacks"`
	RotateBacks  bool         `json:"rotateBacks"`
}

// Print sends the sheets to the printer. Without CUPS (Windows) it opens the
// PDF in the default viewer instead, and reports that with viewerOpened.
func (a *App) Print(req PrintRequest) (viewerOpened bool, err error) {
	a.mu.Lock()
	doc := a.doc
	a.cfg.Printer = req.Printer
	a.cfg.ReverseBacks = req.ReverseBacks
	a.cfg.RotateBacks = req.RotateBacks
	saveConfig(a.cfg)
	a.mu.Unlock()
	if doc == nil {
		return false, errors.New("no file is open")
	}

	opts := impose.Options{Sides: req.Pass}
	if req.Pass == impose.BacksOnly {
		opts.ReverseOrder = req.ReverseBacks
		opts.RotateBacks = req.RotateBacks
	}
	path, err := a.tempOutput(doc, string(req.Pass))
	if err != nil {
		return false, err
	}
	if err := a.writeFile(doc, req.Settings, opts, path); err != nil {
		return false, err
	}

	if !printing.DirectSupported() {
		return true, printing.Open(path)
	}
	defer os.Remove(path)
	return false, printing.Print(printing.Job{
		Printer:  req.Printer,
		File:     path,
		Media:    printing.MediaName(req.Settings.PaperID),
		TwoSided: req.Pass == impose.AllSides && req.Settings.Duplex == book.DuplexAuto,
	})
}

// OpenInViewer writes the print-ready PDF to a temporary file and opens it
// in the system PDF viewer.
func (a *App) OpenInViewer(s book.Settings) error {
	a.mu.Lock()
	doc := a.doc
	a.mu.Unlock()
	if doc == nil {
		return errors.New("no file is open")
	}
	path, err := a.tempOutput(doc, "print")
	if err != nil {
		return err
	}
	if err := a.writeFile(doc, s, impose.Options{Sides: impose.AllSides}, path); err != nil {
		return err
	}
	return printing.Open(path)
}

func (a *App) tempOutput(doc *job.Document, suffix string) (string, error) {
	dir := filepath.Join(os.TempDir(), "book-printer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	base := strings.TrimSuffix(filepath.Base(doc.Path), filepath.Ext(doc.Path))
	return filepath.Join(dir, fmt.Sprintf("%s - %s.pdf", base, suffix)), nil
}

func (a *App) writeFile(doc *job.Document, s book.Settings, opts impose.Options, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := doc.Write(a.ctx, s, opts, f); err != nil {
		f.Close()
		os.Remove(path)
		return err
	}
	return f.Close()
}
