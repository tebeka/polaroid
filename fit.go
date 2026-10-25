package polaroid

import (
	"image"
	"math"

	"golang.org/x/image/draw"
)

// maxCells is the number of row/column diacritics Kitty defines for Unicode placeholders.
const maxCells = 297

// fit returns the largest cell box, within maxCols x maxRows, that shows an imgW x imgH pixel image
// without distorting it, given the pixel size of a single cell.
func fit(imgW, imgH, cellW, cellH, maxCols, maxRows int) (cols, rows int) {
	if imgW <= 0 || imgH <= 0 || maxCols <= 0 || maxRows <= 0 {
		return 0, 0
	}

	maxCols, maxRows = min(maxCols, maxCells), min(maxRows, maxCells)
	scale := min(
		float64(maxCols*cellW)/float64(imgW),
		float64(maxRows*cellH)/float64(imgH),
	)
	cols = int(math.Round(float64(imgW) * scale / float64(cellW)))
	rows = int(math.Round(float64(imgH) * scale / float64(cellH)))
	return max(1, cols), max(1, rows)
}

// shrink scales img down to fit in maxW x maxH pixels, keeping its aspect ratio.
// Images that already fit are returned as is, the terminal scales them up.
func shrink(img image.Image, maxW, maxH int) image.Image {
	b := img.Bounds()
	if b.Dx() <= maxW && b.Dy() <= maxH {
		return img
	}

	scale := min(float64(maxW)/float64(b.Dx()), float64(maxH)/float64(b.Dy()))
	w := max(1, int(math.Round(float64(b.Dx())*scale)))
	h := max(1, int(math.Round(float64(b.Dy())*scale)))
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
	return dst
}
