package polaroid

import (
	"fmt"
	"image"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

// encode returns the sequence that uploads img under id and creates a virtual placement of cols x
// rows cells, which [placeholders] then shows on screen.
func encode(img image.Image, id, cols, rows int) (string, error) {
	opts := kitty.Options{
		Action:           kitty.TransmitAndPut,
		Quite:            2, // suppress all responses
		ID:               id,
		Format:           kitty.PNG,    // compressed, and with straight alpha as Kitty expects (RGBA would be premultiplied)
		Transmission:     kitty.Direct, // EncodeGraphics sends no data when this is left as zero
		Chunk:            true,
		VirtualPlacement: true,
		Columns:          cols,
		Rows:             rows,
	}

	var sb strings.Builder
	if err := kitty.EncodeGraphics(&sb, img, &opts); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// remove returns the sequence that frees the image id from terminal memory.
func remove(id int) string {
	opts := kitty.Options{Action: kitty.Delete, Delete: kitty.DeleteID, DeleteResources: true, ID: id, Quite: 2}
	return ansi.KittyGraphics(nil, opts.Options()...)
}

// placeholders returns cols x rows cells of Unicode placeholders for image id.
// The image ID is encoded in the foreground color, the cell position in the diacritics.
func placeholders(id, cols, rows int) string {
	color := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", (id>>16)&0xFF, (id>>8)&0xFF, id&0xFF)
	var sb strings.Builder
	for row := range rows {
		if row > 0 {
			sb.WriteByte('\n')
		}

		sb.WriteString(color)
		for col := range cols {
			sb.WriteRune(kitty.Placeholder)
			sb.WriteRune(kitty.Diacritic(row))
			sb.WriteRune(kitty.Diacritic(col))
		}

		sb.WriteString("\x1b[39m")
	}

	return sb.String()
}
