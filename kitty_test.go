package polaroid

import (
	"encoding/base64"
	"image"
	"image/png"
	"math/rand/v2"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

func TestEncode(t *testing.T) {
	seq, err := encode(image.NewRGBA(image.Rect(0, 0, 4, 2)), 7, 3, 2)
	if err != nil {
		t.Fatal(err)
	}

	assertOpts(t, seq, "a=T", "U=1", "i=7", "c=3", "r=2", "f=100", "q=2")
}

// Kitty graphics sequence, e.g. "\x1b_Gq=2,m=1;AAAA\x1b\\"
var graphicsRe = regexp.MustCompile(`\x1b_G([^;\x1b]*)(?:;([^\x1b]*))?\x1b\\`)

func TestEncodePayload(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 300, 200))
	rnd := rand.New(rand.NewPCG(1, 2)) // noise doesn't compress, so we need several chunks
	for i := range img.Pix {
		img.Pix[i] = byte(rnd.IntN(256))
		if i%4 == 3 {
			img.Pix[i] = 0xFF // opaque, so premultiplied and straight alpha are the same
		}
	}

	seq, err := encode(img, 7, 3, 2)
	if err != nil {
		t.Fatal(err)
	}

	var payload strings.Builder
	matches := graphicsRe.FindAllStringSubmatch(seq, -1)
	for _, m := range matches {
		payload.WriteString(m[2])
	}

	if len(matches) < 2 {
		t.Fatalf("got %d chunks, want several", len(matches))
	}

	assertOpts(t, matches[0][1], "a=T", "U=1", "i=7", "c=3", "r=2")
	for i, m := range matches {
		want := "m=1" // more chunks follow
		if i == len(matches)-1 {
			want = "m=0"
		}

		assertOpts(t, m[1], want)
	}

	got, err := png.Decode(base64.NewDecoder(base64.StdEncoding, strings.NewReader(payload.String())))
	if err != nil {
		t.Fatal(err)
	}

	if got.Bounds() != img.Bounds() {
		t.Fatalf("bounds = %v, want %v", got.Bounds(), img.Bounds())
	}

	for y := range 200 {
		for x := range 300 {
			if got.At(x, y) != img.At(x, y) {
				t.Fatalf("pixel (%d, %d) = %v, want %v", x, y, got.At(x, y), img.At(x, y))
			}
		}
	}
}

func TestRemove(t *testing.T) {
	assertOpts(t, remove(7), "a=d", "d=I", "i=7", "q=2")
}

func TestPlaceholders(t *testing.T) {
	out := placeholders(0x010203, 3, 2)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}

	for _, line := range lines {
		if w := ansi.StringWidth(line); w != 3 {
			t.Errorf("line width = %d, want 3", w)
		}

		if !strings.HasPrefix(line, "\x1b[38;2;1;2;3m") {
			t.Errorf("line %q doesn't start with the image ID color", line)
		}
	}

	first := string([]rune{kitty.Placeholder, kitty.Diacritic(0), kitty.Diacritic(0)})
	if !strings.Contains(lines[0], first) {
		t.Errorf("first cell %q not found", first)
	}

	last := string([]rune{kitty.Placeholder, kitty.Diacritic(1), kitty.Diacritic(2)})
	if !strings.Contains(lines[1], last) {
		t.Errorf("last cell %q not found", last)
	}
}

func assertOpts(t *testing.T, seq string, opts ...string) {
	t.Helper()
	for _, opt := range opts {
		if !strings.Contains(seq, opt) {
			t.Errorf("missing %q in %q", opt, seq)
		}
	}
}
