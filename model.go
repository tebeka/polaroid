// Package polaroid is a Bubble Tea component that shows images on terminals supporting the Kitty
// graphics protocol (kitty, ghostty).
//
// The image is sent to the terminal out of band, and View renders Kitty Unicode placeholders where
// the terminal draws it. To Bubble Tea these are regular text cells, so layout and redraws work as
// usual.
package polaroid

import (
	"fmt"
	"image"
	"math/rand/v2"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// Model shows an image in a width x height cell box.
type Model struct {
	img image.Image
	// id names the image in terminal memory, the placeholders View renders carry it to tell the terminal
	// which image to draw. It's random so images of other models or programs don't overwrite ours.
	id            int
	width, height int
	cols, rows    int
	cellW, cellH  int
	supported     bool
}

// New returns a Model of width x height cells. Terminal support is detected from the environment;
// use [Model.SetSupported] to override it.
func New(width, height int) Model {
	return Model{
		id:        rand.IntN(1<<24-1) + 1, // the ID is sent as a 24-bit color, 0 is invalid
		width:     width,
		height:    height,
		cellW:     10, // until the terminal reports its cell size
		cellH:     20,
		supported: Supported(os.Environ()),
	}
}

// SetSupported overrides terminal support detection, the returned command sends the image if there is one.
func (m *Model) SetSupported(supported bool) tea.Cmd {
	m.supported = supported
	return m.send()
}

// SetImage sets the image to show, the returned command sends it to the terminal.
func (m *Model) SetImage(img image.Image) tea.Cmd {
	m.img = img
	m.refit()
	return m.send()
}

// SetSize sets the box size in cells. If the image size changes, the returned command sends it again.
func (m *Model) SetSize(width, height int) tea.Cmd {
	m.width, m.height = width, height
	if !m.refit() {
		return nil
	}

	return m.send()
}

// Init asks the terminal for its cell size, needed to keep the image aspect ratio, and sends the image if
// one was set before the program started.
func (m Model) Init() tea.Cmd {
	if !m.supported {
		return nil
	}

	return tea.Batch(tea.Raw(ansi.WindowOp(16)), m.send()) // 16: report cell size in pixels
}

// Update handles the terminal cell size report.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	size, ok := msg.(uv.CellSizeEvent)
	if !ok || size.Width <= 0 || size.Height <= 0 || (size.Width == m.cellW && size.Height == m.cellH) {
		return m, nil
	}

	// Send even if cols & rows didn't change, the image pixel size depends on the cell size.
	m.cellW, m.cellH = size.Width, size.Height
	m.refit()
	return m, m.send()
}

// View renders the image placeholders at the top left of a width x height box.
func (m Model) View() string {
	if !m.supported && m.cols > 0 { // refit sets cols to 0 when there's no image
		return message("image rendering not supported", m.width, m.height)
	}

	if !m.visible() {
		return blank(m.width, m.height)
	}

	var sb strings.Builder
	pad := strings.Repeat(" ", m.width-m.cols)
	for i, line := range strings.Split(placeholders(m.id, m.cols, m.rows), "\n") {
		if i > 0 {
			sb.WriteByte('\n')
		}

		sb.WriteString(line)
		sb.WriteString(pad)
	}

	if m.height > m.rows {
		sb.WriteByte('\n')
		sb.WriteString(blank(m.width, m.height-m.rows))
	}

	return sb.String()
}

// Clear returns a command that frees the image from terminal memory, call it before quitting.
func (m Model) Clear() tea.Cmd {
	if !m.supported {
		return nil
	}

	return tea.Raw(remove(m.id))
}

// refit recalculates the image size in cells and reports whether it changed.
func (m *Model) refit() bool {
	var cols, rows int
	if m.img != nil {
		b := m.img.Bounds()
		cols, rows = fit(b.Dx(), b.Dy(), m.cellW, m.cellH, m.width, m.height)
	}

	changed := cols != m.cols || rows != m.rows
	m.cols, m.rows = cols, rows
	return changed
}

func (m Model) visible() bool {
	return m.supported && m.cols > 0 // refit sets cols to 0 when there's no image
}

func (m Model) send() tea.Cmd {
	if !m.visible() {
		return nil
	}

	// Encode here so the command only writes a ready string. Commands run concurrently, this keeps the window
	// for frames arriving out of order negligible.
	img := shrink(m.img, m.cols*m.cellW, m.rows*m.cellH)
	seq, err := encode(img, m.id, m.cols, m.rows)
	if err != nil {
		return func() tea.Msg { return err }
	}

	return tea.Raw(seq)
}

// message renders msg at the top left of a width x height box, cut to the box width.
func message(msg string, width, height int) string {
	if height <= 0 {
		return ""
	}

	if height == 1 {
		return fmt.Sprintf("%-*.*s", width, width, msg)
	}

	return fmt.Sprintf("%-*.*s\n%s", width, width, msg, blank(width, height-1))
}

func blank(width, height int) string {
	line := strings.Repeat(" ", max(width, 0))
	return strings.TrimSuffix(strings.Repeat(line+"\n", max(height, 0)), "\n")
}
