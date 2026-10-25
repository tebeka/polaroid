package polaroid

import (
	"image"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

func TestViewNoImage(t *testing.T) {
	m := New(5, 3)
	m.SetSupported(true)
	assertBox(t, m.View(), 5, 3)
}

func TestViewUnsupported(t *testing.T) {
	m := New(5, 3)
	m.SetSupported(false)
	if cmd := m.SetImage(image.NewRGBA(image.Rect(0, 0, 100, 100))); cmd != nil {
		t.Error("SetImage returned a command on an unsupported terminal")
	}

	assertBox(t, m.View(), 5, 3)
	if strings.Contains(m.View(), "\x1b") {
		t.Error("View has escape sequences on an unsupported terminal")
	}

	if !strings.HasPrefix(m.View(), "image\n") { // message cut to the box width
		t.Errorf("View = %q, want the unsupported message", m.View())
	}
}

func TestViewUnsupportedNoImage(t *testing.T) {
	m := New(5, 3)
	m.SetSupported(false)
	if v := m.View(); strings.TrimSpace(v) != "" {
		t.Errorf("View = %q, want blank", v)
	}
}

func TestViewImage(t *testing.T) {
	m := New(80, 20)
	m.SetSupported(true)
	if cmd := m.SetImage(image.NewRGBA(image.Rect(0, 0, 100, 100))); cmd == nil {
		t.Fatal("SetImage returned no command")
	}

	// 100x100 image with 10x20 cells in 80x20 → 40x20, padded to the box
	assertBox(t, m.View(), 80, 20)
}

func TestViewPadsRows(t *testing.T) {
	m := New(80, 20)
	m.SetSupported(true)
	m.SetImage(image.NewRGBA(image.Rect(0, 0, 400, 100))) // 80x10
	assertBox(t, m.View(), 80, 20)
}

func TestSetSupportedSends(t *testing.T) {
	m := New(80, 20)
	m.SetSupported(false)
	m.SetImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))
	if cmd := m.SetSupported(true); cmd == nil {
		t.Error("SetSupported(true) didn't send the image")
	}
}

func TestSetSizeUnchanged(t *testing.T) {
	m := New(80, 20)
	m.SetSupported(true)
	m.SetImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))
	if cmd := m.SetSize(80, 20); cmd != nil {
		t.Error("SetSize returned a command when nothing changed")
	}

	if cmd := m.SetSize(40, 10); cmd == nil {
		t.Error("SetSize returned no command after a resize")
	}
}

func TestCellSize(t *testing.T) {
	m := New(80, 20)
	m.SetSupported(true)
	m.SetImage(image.NewRGBA(image.Rect(0, 0, 1000, 1000)))

	m, cmd := m.Update(uv.CellSizeEvent{Width: 20, Height: 40}) // same aspect ratio, same cols & rows
	if cmd == nil {
		t.Error("no command after a cell size change")
	}

	if _, cmd = m.Update(uv.CellSizeEvent{Width: 20, Height: 40}); cmd != nil {
		t.Error("command for an unchanged cell size")
	}
}

func assertBox(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("got %d lines, want %d", len(lines), height)
	}

	for _, line := range lines {
		if w := ansi.StringWidth(line); w != width {
			t.Errorf("line width = %d, want %d", w, width)
		}
	}
}
