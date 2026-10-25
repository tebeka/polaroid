# polaroid

<p align="center"><img src="example/logo.png" width="300" alt="polaroid logo"></p>

A [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2 component that shows images on terminals
supporting the [Kitty graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/)
([kitty](https://sw.kovidgoyal.net/kitty/), [ghostty](https://ghostty.org/)). (Kitty, Ghostty currently)
Inspired by [textual-image](https://github.com/lnqs/textual-image).

## Usage

```go
type model struct {
	pic polaroid.Model // polaroid.New(40, 20), width & height in cells
}

func (m model) Init() tea.Cmd {
	return m.pic.Init() // asks the terminal for its cell size
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case imageMsg:
		return m, m.pic.SetImage(msg.img) // any image.Image
	case tea.WindowSizeMsg:
		return m, m.pic.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		return m, tea.Sequence(m.pic.Clear(), tea.Quit) // free terminal memory on quit
	}

	var cmd tea.Cmd
	m.pic, cmd = m.pic.Update(msg) // handles the cell size report
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView(m.pic.View())
}
```

The image keeps its aspect ratio and is scaled, up or down, to the largest size that fits the box. The terminal
enlarges images smaller than the box, so these may look blurry.

Support is detected from the environment (`TERM`, `TERM_PROGRAM`, `KITTY_WINDOW_ID`); override it with
`SetSupported`. On unsupported terminals `View` renders "image rendering not supported" in the box.

When serving over SSH (e.g. with [wish](https://github.com/charmbracelet/wish)), the process environment
describes the server, not the client. Detect support per session with
`m.pic.SetSupported(polaroid.Supported(sess.Environ()))`.

The image ID is sent as a 24-bit color, so Bubble Tea must use true color. This is automatic with
`TERM=xterm-kitty` or `TERM=xterm-ghostty`; with any other `TERM` (e.g. over SSH) set `COLORTERM=truecolor`.

## Example


```go
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"log"

	tea "charm.land/bubbletea/v2"

	"github.com/tebeka/polaroid"
)

//go:embed logo.png
var logoPNG []byte

const (
	maxZoom  = 8.0
	zoomStep = 1.25
	panStep  = 0.1 // fraction of the visible area
)

// subImager is an image that can be cropped, all stdlib image types are.
type subImager interface {
	image.Image
	SubImage(image.Rectangle) image.Image
}

type model struct {
	src    subImager
	pic    polaroid.Model
	zoom   float64
	cx, cy float64 // view center, as a fraction of the image size
}

func (m model) Init() tea.Cmd {
	return m.pic.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, m.pic.SetSize(msg.Width, msg.Height-1) // leave room for the status line
	case tea.KeyPressMsg:
		return m.handleKey(msg.String())
	}

	var cmd tea.Cmd
	m.pic, cmd = m.pic.Update(msg)
	return m, cmd
}

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	prev := m
	step := panStep / m.zoom
	switch key {
	case "q", "ctrl+c":
		return m, tea.Sequence(m.pic.Clear(), tea.Quit)
	case "+", "=":
		m.zoom = min(m.zoom*zoomStep, maxZoom)
	case "-":
		m.zoom = max(m.zoom/zoomStep, 1)
	case "0":
		m.zoom, m.cx, m.cy = 1, 0.5, 0.5
	case "left", "h":
		m.cx -= step
	case "right", "l":
		m.cx += step
	case "up", "k":
		m.cy -= step
	case "down", "j":
		m.cy += step
	default:
		return m, nil
	}

	m.cx, m.cy = clampCenter(m.cx, m.zoom), clampCenter(m.cy, m.zoom)
	if m.zoom == prev.zoom && m.cx == prev.cx && m.cy == prev.cy { // e.g. panning at the edge
		return m, nil
	}
	return m, m.pic.SetImage(m.src.SubImage(crop(m.src.Bounds(), m.zoom, m.cx, m.cy)))
}

func (m model) View() tea.View {
	status := fmt.Sprintf("zoom %.1fx · +/- zoom · arrows/hjkl pan · 0 reset · q quit", m.zoom)
	v := tea.NewView(m.pic.View() + "\n" + status)
	v.AltScreen = true
	return v
}

// crop returns the part of bounds visible at zoom, centered at (cx, cy) given as fractions of bounds.
// The center should be clamped with [clampCenter].
func crop(bounds image.Rectangle, zoom, cx, cy float64) image.Rectangle {
	// Size first so it doesn't change with the center, a jittering size would change the image aspect ratio.
	w, h := int(float64(bounds.Dx())/zoom), int(float64(bounds.Dy())/zoom)
	x := int(cx*float64(bounds.Dx())) - w/2
	y := int(cy*float64(bounds.Dy())) - h/2
	x, y = max(0, min(x, bounds.Dx()-w)), max(0, min(y, bounds.Dy()-h))
	return image.Rect(x, y, x+w, y+h).Add(bounds.Min)
}

// clampCenter keeps the view center far enough from the edges for the view to stay inside the image.
func clampCenter(c, zoom float64) float64 {
	half := 0.5 / zoom
	return max(half, min(c, 1-half))
}

func main() {
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		log.Fatal(err)
	}
	src := img.(subImager)

	pic := polaroid.New(0, 0)
	pic.SetImage(src) // no command until we get the window size
	m := model{src: src, pic: pic, zoom: 1, cx: 0.5, cy: 0.5}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatal(err)
	}
}
```

Keys: `+`/`-` zoom, arrows or `hjkl` pan, `0` reset, `q` quit.

## Limitations

- Kitty graphics protocol only, no sixel or text fallback.
- Doesn't work inside tmux.
