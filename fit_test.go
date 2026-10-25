package polaroid

import (
	"image"
	"testing"
)

func TestFit(t *testing.T) {
	cases := []struct {
		name               string
		imgW, imgH         int
		maxCols, maxRows   int
		wantCols, wantRows int
	}{
		// cells are 10x20 pixels
		{"square in wide box", 100, 100, 80, 20, 40, 20},
		{"wide image", 400, 100, 80, 40, 80, 10},
		{"tall image", 100, 400, 80, 20, 10, 20},
		{"tiny box", 1000, 1000, 1, 1, 1, 1},
		{"cap at diacritics", 10000, 100, 1000, 1000, 297, 1},
		{"empty image", 0, 0, 80, 20, 0, 0},
		{"empty box", 100, 100, 0, 0, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cols, rows := fit(tc.imgW, tc.imgH, 10, 20, tc.maxCols, tc.maxRows)
			if cols != tc.wantCols || rows != tc.wantRows {
				t.Errorf("fit = %dx%d, want %dx%d", cols, rows, tc.wantCols, tc.wantRows)
			}
		})
	}
}

func TestShrinkFits(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	if got := shrink(img, 400, 400); got != img {
		t.Error("shrink returned a new image for one that fits")
	}
}

func TestShrink(t *testing.T) {
	cases := []struct {
		name         string
		w, h         int
		maxW, maxH   int
		wantW, wantH int
	}{
		{"wide", 2000, 1000, 400, 400, 400, 200},
		{"tall", 1000, 2000, 400, 400, 200, 400},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
			b := shrink(img, tc.maxW, tc.maxH).Bounds()
			if b.Dx() != tc.wantW || b.Dy() != tc.wantH {
				t.Errorf("shrink = %dx%d, want %dx%d", b.Dx(), b.Dy(), tc.wantW, tc.wantH)
			}
		})
	}
}
