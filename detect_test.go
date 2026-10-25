package polaroid

import "testing"

func TestSupported(t *testing.T) {
	cases := []struct {
		name    string
		environ []string
		want    bool
	}{
		{"kitty term", []string{"TERM=xterm-kitty"}, true},
		{"ghostty term", []string{"TERM=xterm-ghostty"}, true},
		{"ghostty program", []string{"TERM=xterm-256color", "TERM_PROGRAM=ghostty", "COLORTERM=truecolor"}, true},
		{"kitty window", []string{"TERM=xterm-256color", "KITTY_WINDOW_ID=1", "COLORTERM=truecolor"}, true},
		{"kitty window, no true color", []string{"TERM=xterm-256color", "KITTY_WINDOW_ID=1"}, false},
		{"ghostty program, no true color", []string{"TERM=xterm-256color", "TERM_PROGRAM=ghostty"}, false},
		{"kitty term, no color", []string{"TERM=xterm-kitty", "NO_COLOR=1"}, false},
		{"xterm", []string{"TERM=xterm-256color", "COLORTERM=truecolor"}, false},
		{"tmux in kitty", []string{"TERM=tmux-256color", "KITTY_WINDOW_ID=1", "COLORTERM=truecolor", "TMUX=/tmp/tmux-1000/default,1,0"}, false},
		{"screen in ghostty", []string{"TERM=screen", "TERM_PROGRAM=ghostty", "COLORTERM=truecolor", "STY=1.pts-0.host"}, false},
		{"empty", nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Supported(tc.environ); got != tc.want {
				t.Errorf("Supported(%v) = %v, want %v", tc.environ, got, tc.want)
			}
		})
	}
}
