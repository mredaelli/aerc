	"fmt"
	"github.com/danwakefield/fnmatch"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/google/shlex"
	"git.sr.ht/~sircmpwn/aerc2/config"
	"git.sr.ht/~sircmpwn/aerc2/lib"
	"git.sr.ht/~sircmpwn/aerc2/worker/types"
	conf    *config.AercConfig
	err     error
	filter  *exec.Cmd
	msg     *types.MessageInfo
	pager   *exec.Cmd
	source  io.Reader
	pagerin io.WriteCloser
	sink    io.WriteCloser
	grid    *ui.Grid
	term    *Terminal
func formatAddresses(addrs []*imap.Address) string {
	val := bytes.Buffer{}
	for i, addr := range addrs {
		if addr.PersonalName != "" {
			val.WriteString(fmt.Sprintf("%s <%s@%s>",
				addr.PersonalName, addr.MailboxName, addr.HostName))
		} else {
			val.WriteString(fmt.Sprintf("%s@%s",
				addr.MailboxName, addr.HostName))
		}
		if i != len(addrs)-1 {
			val.WriteString(", ")
		}
	}
	return val.String()
}

func NewMessageViewer(conf *config.AercConfig, store *lib.MessageStore,
	msg *types.MessageInfo) *MessageViewer {

		{ui.SIZE_EXACT, 3}, // TODO: Based on number of header rows
	// TODO: let user specify additional headers to show by default
			Value: formatAddresses(msg.Envelope.From),
			Value: formatAddresses(msg.Envelope.To),
			Name:  "Subject",
			Value: msg.Envelope.Subject,
	headers.AddChild(ui.NewFill(' ')).At(2, 0).Span(1, 2)
	var (
		filter  *exec.Cmd
		pager   *exec.Cmd
		pipe    io.WriteCloser
		pagerin io.WriteCloser
		term    *Terminal
		viewer  *MessageViewer
	)
	cmd, err := shlex.Split(conf.Viewer.Pager)
	if err != nil {
		goto handle_error
	}
	pager = exec.Command(cmd[0], cmd[1:]...)

	for _, f := range conf.Filters {
		mime := msg.BodyStructure.MIMEType + "/" + msg.BodyStructure.MIMESubType
		switch f.FilterType {
		case config.FILTER_MIMETYPE:
			if fnmatch.Match(f.Filter, mime, 0) {
				filter = exec.Command("sh", "-c", f.Command)
			}
		case config.FILTER_HEADER:
			var header string
			switch f.Header {
			case "subject":
				header = msg.Envelope.Subject
			case "from":
				header = formatAddresses(msg.Envelope.From)
			case "to":
				header = formatAddresses(msg.Envelope.To)
			case "cc":
				header = formatAddresses(msg.Envelope.Cc)
			}
			if f.Regex.Match([]byte(header)) {
				filter = exec.Command("sh", "-c", f.Command)
			}
		}
		if filter != nil {
			break
		}
	}
	if filter != nil {
		pipe, _ = filter.StdinPipe()
		pagerin, _ = pager.StdinPipe()
	} else {
		pipe, _ = pager.StdinPipe()
	term, _ = NewTerminal(pager)
	// TODO: configure multipart view. I left a spot for it in the grid
	body.AddChild(term).At(0, 0).Span(1, 2)

	viewer = &MessageViewer{
		filter:  filter,
		grid:    grid,
		msg:     msg,
		pager:   pager,
		pagerin: pagerin,
		sink:    pipe,
		term:    term,
	}

	store.FetchBodyPart(msg.Uid, 0, func(reader io.Reader) {
		viewer.source = reader
		viewer.attemptCopy()
	})

	term.OnStart = func() {
		viewer.attemptCopy()
	}

	return viewer

handle_error:
	viewer = &MessageViewer{
		err:  err,
		grid: grid,
		msg:  msg,
	}
	return viewer
}

func (mv *MessageViewer) attemptCopy() {
	if mv.source != nil && mv.pager.Process != nil {
		header := make(message.Header)
		header.Set("Content-Transfer-Encoding", mv.msg.BodyStructure.Encoding)
		header.SetContentType(
			mv.msg.BodyStructure.MIMEType, mv.msg.BodyStructure.Params)
		header.SetContentDescription(mv.msg.BodyStructure.Description)
		if mv.filter != nil {
			stdout, _ := mv.filter.StdoutPipe()
			mv.filter.Start()
			go func() {
				_, err := io.Copy(mv.pagerin, stdout)
				if err != nil {
					mv.err = err
					mv.Invalidate()
				}
				mv.pagerin.Close()
				stdout.Close()
			}()
		}
		go func() {
			entity, err := message.New(header, mv.source)
			if err != nil {
				mv.err = err
				mv.Invalidate()
				return
			}
			reader := mail.NewReader(entity)
			part, err := reader.NextPart()
			if err != nil {
				mv.err = err
				mv.Invalidate()
				return
			}
			io.Copy(mv.sink, part.Body)
			mv.sink.Close()
		}()
	}
	if mv.err != nil {
		ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', tcell.StyleDefault)
		ctx.Printf(0, 0, tcell.StyleDefault, "%s", mv.err.Error())
		return
	}
	if mv.term != nil {
		return mv.term.Event(event)
	}
	return false
	if mv.term != nil {
		mv.term.Focus(focus)
	}
	name := hv.Name
	size := runewidth.StringWidth(name)
	lim := ctx.Width() - size - 1
	value := runewidth.Truncate(" "+hv.Value, lim, "…")
	ctx.Printf(0, 0, hstyle, name)
	ctx.Printf(size, 0, vstyle, value)