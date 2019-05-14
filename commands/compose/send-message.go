package compose

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/gdamore/tcell"

	"git.sr.ht/~sircmpwn/aerc2/widgets"
)

func init() {
	register("send-message", SendMessage)
}

func SendMessage(aerc *widgets.Aerc, args []string) error {
	if len(args) > 1 {
		return errors.New("Usage: send-message")
	}
	composer, _ := aerc.SelectedTab().(*widgets.Composer)
	config := composer.Config()

	if config.Outgoing == "" {
		return errors.New(
			"No outgoing mail transport configured for this account")
	}

	uri, err := url.Parse(config.Outgoing)
	if err != nil {
		return err
	}
	var (
		scheme string
		auth   string = "plain"
	)
	parts := strings.Split(uri.Scheme, "+")
	if len(parts) == 1 {
		scheme = parts[0]
	} else if len(parts) == 2 {
		scheme = parts[0]
		auth = parts[1]
	} else {
		return fmt.Errorf("Unknown transfer protocol %s", uri.Scheme)
	}

	header, rcpts, err := composer.Header()
	if err != nil {
		return err
	}

	if config.From == "" {
		return errors.New("No 'From' configured for this account")
	}
	from, err := mail.ParseAddress(config.From)
	if err != nil {
		return err
	}

	var (
		saslClient sasl.Client
		conn       *smtp.Client
	)
	switch auth {
	case "":
		fallthrough
	case "none":
		saslClient = nil
	case "plain":
		password, _ := uri.User.Password()
		saslClient = sasl.NewPlainClient("", uri.User.Username(), password)
	default:
		return fmt.Errorf("Unsupported auth mechanism %s", auth)
	}

	aerc.SetStatus("Sending...")

	sendAsync := func() {
		tlsConfig := &tls.Config{
			// TODO: ask user first
			InsecureSkipVerify: true,
		}
		switch scheme {
		case "smtp":
			host := uri.Host
			if !strings.ContainsRune(host, ':') {
				host = host + ":587" // Default to submission port
			}
			conn, err = smtp.Dial(host)
			if err != nil {
				aerc.PushStatus(" "+err.Error(), 10*time.Second).
					Color(tcell.ColorDefault, tcell.ColorRed)
				return
			}
			defer conn.Close()
			if sup, _ := conn.Extension("STARTTLS"); sup {
				// TODO: let user configure tls?
				if err = conn.StartTLS(tlsConfig); err != nil {
					aerc.PushStatus(" "+err.Error(), 10*time.Second).
						Color(tcell.ColorDefault, tcell.ColorRed)
					return
				}
			}
		case "smtps":
			host := uri.Host
			if !strings.ContainsRune(host, ':') {
				host = host + ":465" // Default to smtps port
			}
			conn, err = smtp.DialTLS(host, tlsConfig)
			if err != nil {
				aerc.PushStatus(" "+err.Error(), 10*time.Second).
					Color(tcell.ColorDefault, tcell.ColorRed)
				return
			}
			defer conn.Close()
		}

		// TODO: sendmail
		if saslClient != nil {
			if err = conn.Auth(saslClient); err != nil {
				aerc.PushStatus(" "+err.Error(), 10*time.Second).
					Color(tcell.ColorDefault, tcell.ColorRed)
				return
			}
		}
		// TODO: the user could conceivably want to use a different From and sender
		if err = conn.Mail(from.Address); err != nil {
			aerc.PushStatus(" "+err.Error(), 10*time.Second).
				Color(tcell.ColorDefault, tcell.ColorRed)
			return
		}
		for _, rcpt := range rcpts {
			if err = conn.Rcpt(rcpt); err != nil {
				aerc.PushStatus(" "+err.Error(), 10*time.Second).
					Color(tcell.ColorDefault, tcell.ColorRed)
				return
			}
		}
		wc, err := conn.Data()
		if err != nil {
			aerc.PushStatus(" "+err.Error(), 10*time.Second).
				Color(tcell.ColorDefault, tcell.ColorRed)
			return
		}
		defer wc.Close()
		composer.WriteMessage(header, wc)
		composer.Close()
		aerc.RemoveTab(composer)
	}

	go func() {
		sendAsync()
		// TODO: Use a stack
		aerc.SetStatus("Sent.")
	}()
	return nil
}
