package main

import (
	"image"
	"testing"
)

func TestCrop(t *testing.T) {
	bounds := image.Rect(0, 0, 100, 200)
	cases := []struct {
		name   string
		zoom   float64
		cx, cy float64
		want   image.Rectangle
	}{
		{"full", 1, 0.5, 0.5, bounds},
		{"zoom 2 centered", 2, 0.5, 0.5, image.Rect(25, 50, 75, 150)},
		{"zoom 2 top left", 2, 0.25, 0.25, image.Rect(0, 0, 50, 100)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := crop(bounds, tc.zoom, tc.cx, tc.cy); got != tc.want {
				t.Errorf("crop = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCropStableSize(t *testing.T) {
	bounds := image.Rect(0, 0, 1024, 1169)
	zoom := 1.25 * 1.25 * 1.25
	want := crop(bounds, zoom, 0.5, 0.5).Size()
	for cx := 0.32; cx < 0.68; cx += 0.01 {
		if got := crop(bounds, zoom, cx, 0.5).Size(); got != want {
			t.Fatalf("crop size at cx=%v = %v, want %v", cx, got, want)
		}
	}
}

func TestCropOffsetBounds(t *testing.T) {
	bounds := image.Rect(10, 20, 110, 220)
	if got, want := crop(bounds, 2, 0.5, 0.5), image.Rect(35, 70, 85, 170); got != want {
		t.Errorf("crop = %v, want %v", got, want)
	}
}

func TestClampCenter(t *testing.T) {
	cases := []struct {
		name    string
		c, zoom float64
		want    float64
	}{
		{"inside", 0.4, 2, 0.4},
		{"low edge", 0, 2, 0.25},
		{"high edge", 1, 2, 0.75},
		{"full view", 0.9, 1, 0.5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampCenter(tc.c, tc.zoom); got != tc.want {
				t.Errorf("clampCenter(%v, %v) = %v, want %v", tc.c, tc.zoom, got, tc.want)
			}
		})
	}
}
