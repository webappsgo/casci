package cli

import (
	"os"
	"strings"
)

// ColorEnabled reports whether ANSI color should be emitted.
// Priority: the --color flag, then NO_COLOR, then TTY and TERM auto-detection.
func ColorEnabled(flagValue string) bool {
	switch strings.ToLower(flagValue) {
	case "yes":
		return true
	case "no":
		return false
	}

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return isTerminal(os.Stdout)
}

// EmojiEnabled reports whether emoji should appear in terminal output.
// NO_COLOR and TERM=dumb both request plain output.
func EmojiEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

// isTerminal reports whether f refers to a character device.
// Checking the file mode avoids pulling in a terminal dependency for what is
// only ever used to decide whether to colorize.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// icon returns the emoji when emoji output is enabled, or a plain-text marker
// so that NO_COLOR output stays readable.
func icon(emoji, plain string) string {
	if EmojiEnabled() {
		return emoji
	}
	return plain
}
