package commands

import (
	"errors"
	"git.sr.ht/~sircmpwn/getopt"
	"time"

	"git.sr.ht/~sircmpwn/aerc/widgets"
)

type Quit struct{}

func init() {
	register(Quit{})
}

func (_ Quit) Aliases() []string {
	return []string{"quit", "exit"}
}

func (_ Quit) Complete(aerc *widgets.Aerc, args []string) []string {
	return nil
}

type ErrorExit int

func (err ErrorExit) Error() string {
	return "exit"
}

func (_ Quit) Execute(aerc *widgets.Aerc, args []string) error {
	opts, optind, err := getopt.Getopts(args, "y")
	if err != nil {
		return err
	}
	var dontAsk bool
	for _, opt := range opts {
		switch opt.Option {
		case 'y':
			dontAsk = true
		}
	}
	cmd := args[optind:]
	if len(cmd) > 0 {
		return errors.New("Usage: quit [-y]")
	}
	if !dontAsk {
		for _, name := range aerc.TabNames() {
			if !aerc.CanCloseTab(name) {
				aerc.SelectTab(name)
				aerc.PushStatus("This tab has unsaved changes.", 5*time.Second)
			}
		}
	}
	return ErrorExit(1)
}
