package menu

import (
	"os"

	"github.com/chzyer/readline"
)

// This is a fix for the annoying promptui bell ringing.
// See https://github.com/manifoldco/promptui/issues/49
// Fix from https://github.com/PremiereGlobal/stim/pull/21

func init() {
	readline.Stdout = &noBellStderr{}
}

// noBellStderr implements an io.WriteCloser that skips the terminal bell character
// (ASCII code 7), and writes the rest to os.Stderr. It's used to replace
// readline.Stdout, that is the package used by promptui to display the prompts.
type noBellStderr struct{}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (s *noBellStderr) Write(b []byte) (int, error) {
	if len(b) == 1 && b[0] == readline.CharBell {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (s *noBellStderr) Close() error {
	return os.Stderr.Close()
}
