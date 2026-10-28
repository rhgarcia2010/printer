# Book Printer

Turns a book PDF exported from Canva (one page per book page, in reading order)
into a print-ready PDF of folded sheets, checks it for problems, and sends it to
the printer.

- **Bindings:** stapled booklet (saddle stitch) or signatures for sewn/glued books.
- **Paper:** US Letter, US Legal, A4, A5, each folded in half.
- **Checks:** blank-page padding, mixed page sizes, scaling, content cut off by the
  printer's unprintable margin, content too close to the edge or fold, full-bleed art.
- **Printing:** automatic duplex (short-edge flip) or a guided two-pass manual duplex.
  Uses CUPS (`lp`) on Linux/macOS; on Windows the PDF opens in your PDF viewer.

Settings, the last file and the printer choice are remembered in
`<user config dir>/book-printer/config.json`.

## Build

On macOS or Linux:

```sh
make doctor   # checks prerequisites and says what is missing
make setup    # installs the Wails CLI and frontend packages (first time only)
make build    # builds the app into build/bin
```

`make help` lists the rest (`run`, `dev`, `test`, `check`, `clean`, and
`build-universal` for a Mac app that runs on both Intel and Apple Silicon).

Requirements: Go 1.26+, Node.js 20.19+ (or 22.12+), and the Wails CLI:

```sh
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

| OS | Extra requirements | Command |
|---|---|---|
| Linux | gcc, pkg-config, `gtk3`, `webkit2gtk-4.1` | `wails build -tags webkit2_41` |
| macOS | Xcode command line tools | `wails build` |
| Windows | WebView2 runtime (preinstalled on Windows 10/11) | `wails build` |

The result is a single binary in `build/bin/`. Wails apps link against the
system web view, so build on each OS (or in CI) rather than cross-compiling.

During development, `wails dev -tags webkit2_41` gives a live-reloading window.

You can also pass a PDF on the command line: `book-printer mybook.pdf`.

## Layout

| Package | Role |
|---|---|
| `internal/book` | Paper sizes, settings, the page→sheet plan, placement geometry. Pure logic. |
| `internal/impose` | Writes the imposed PDF with pdfcpu (pages become form XObjects). |
| `internal/raster` | Renders pages with pdfium compiled to WebAssembly (no cgo, no native libs). |
| `internal/preflight` | The checks, phrased for the person printing. |
| `internal/job` | Load → check → preview → write, used by the UI. |
| `internal/printing` | CUPS / default-viewer handoff. |
| `frontend/` | Svelte 5 UI. |

Run the tests with `go test ./...`.

## Adding a binding style

A binding is a way of assigning book pages to sheet sides. `book.NewPlan`
handles everything that folds sheets in half (booklets and signatures). A
binding that does not fold (e.g. single sheets for perfect binding) needs its own
planner and a matching placement in `book.PlaceSide`; the rest of the pipeline
(checks, preview, printing) works per sheet side and does not change.
