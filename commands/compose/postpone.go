package compose

import (
	"fmt"
	// "os"
	"time"

	// "git.sr.ht/~sircmpwn/aerc/commands"
	"git.sr.ht/~sircmpwn/aerc/widgets"
	"github.com/gdamore/tcell"
	// "github.com/mitchellh/go-homedir"
)

type Postpone struct{}

func init() {
	register(Postpone{})
}

func (_ Postpone) Aliases() []string {
	return []string{"postpone"}
}

func (_ Postpone) Complete(aerc *widgets.Aerc, args []string) []string {
	// path := ""
	// if len(args) >= 1 {
	// 	path = args[0]
	// }

	// return commands.CompletePath(path)
	return []string{"ciao"}
}

func (_ Postpone) Execute(aerc *widgets.Aerc, args []string) error {
	// if len(args) != 2 {
	// 	return fmt.Errorf("Usage: :attach <path>")
	// }

	// path := args[1]

	// path, err := homedir.Expand(path)
	// if err != nil {
	// 	aerc.PushError(" " + err.Error())
	// 	return err
	// }

	// pathinfo, err := os.Stat(path)
	// if err != nil {
	// 	aerc.PushError(" " + err.Error())
	// 	return err
	// } else if pathinfo.IsDir() {
	// 	aerc.PushError("Attachment must be a file, not a directory")
	// 	return nil
	// }

	// composer, _ := aerc.SelectedTab().(*widgets.Composer)
	// header, _, _ := composer.PrepareHeader()
	// composer.WriteMessage(header, w)
	// w.Close()

	aerc.PushStatus(fmt.Sprintf("Attached %s", ""), 10*time.Second).
		Color(tcell.ColorDefault, tcell.ColorGreen)

	return nil
}
