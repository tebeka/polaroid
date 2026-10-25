package polaroid

import (
	"slices"
	"strings"

	"github.com/charmbracelet/colorprofile"
)

// Supported reports whether the terminal described by environ (in [os.Environ] format) supports the
// Kitty graphics protocol with Unicode placeholders.
//
// The image ID is sent as a 24-bit color, so the environment must also make Bubble Tea use true color
// (e.g. COLORTERM=truecolor), otherwise it downsamples the color and the image doesn't show.
// Terminal multiplexers (tmux, screen) are not supported, even when running inside kitty or ghostty.
func Supported(environ []string) bool {
	if colorprofile.Env(environ) != colorprofile.TrueColor {
		return false
	}

	inMultiplexer := slices.ContainsFunc(environ, func(kv string) bool {
		return strings.HasPrefix(kv, "TMUX=") || strings.HasPrefix(kv, "STY=")
	})
	if inMultiplexer {
		return false
	}

	return slices.ContainsFunc(environ, func(kv string) bool {
		return kv == "TERM=xterm-kitty" || kv == "TERM=xterm-ghostty" || kv == "TERM_PROGRAM=ghostty" ||
			strings.HasPrefix(kv, "KITTY_WINDOW_ID=")
	})
}
