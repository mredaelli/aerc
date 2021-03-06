package models

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
)

// Flag is an abstraction around the different flags which can be present in
// different email backends and represents a flag that we use in the UI.
type Flag int

const (
	// SeenFlag marks a message as having been seen previously
	SeenFlag Flag = iota

	// RecentFlag marks a message as being recent
	RecentFlag

	// AnsweredFlag marks a message as having been replied to
	AnsweredFlag

	// DeletedFlag marks a message as having been deleted
	DeletedFlag

	// FlaggedFlag marks a message with a user flag
	FlaggedFlag
)

type Directory struct {
	Name       string
	Attributes []string
}

type DirectoryInfo struct {
	Name     string
	Flags    []string
	ReadOnly bool

	// The total number of messages in this mailbox.
	Exists int

	// The number of messages not seen since the last time the mailbox was opened.
	Recent int

	// The number of unread messages
	Unseen int
}

// A MessageInfo holds information about the structure of a message
type MessageInfo struct {
	BodyStructure *BodyStructure
	Envelope      *Envelope
	Flags         []Flag
	InternalDate  time.Time
	RFC822Headers *mail.Header
	Size          uint32
	Uid           uint32
}

// A MessageBodyPart can be displayed in the message viewer
type MessageBodyPart struct {
	Reader io.Reader
	Uid    uint32
}

// A FullMessage is the entire message
type FullMessage struct {
	Reader io.Reader
	Uid    uint32
}

type BodyStructure struct {
	MIMEType          string
	MIMESubType       string
	Params            map[string]string
	Description       string
	Encoding          string
	Parts             []*BodyStructure
	Disposition       string
	DispositionParams map[string]string
}

type Envelope struct {
	Date      time.Time
	Subject   string
	From      []*Address
	ReplyTo   []*Address
	To        []*Address
	Cc        []*Address
	Bcc       []*Address
	MessageId string
}

type Address struct {
	Name    string
	Mailbox string
	Host    string
}

var atom *regexp.Regexp = regexp.MustCompile("^[a-z0-9!#$%7'*+-/=?^_`{}|~ ]+$")

func (a Address) Format() string {
	if a.Name != "" {
		if atom.MatchString(a.Name) {
			return fmt.Sprintf("%s <%s@%s>", a.Name, a.Mailbox, a.Host)
		} else {
			return fmt.Sprintf("\"%s\" <%s@%s>",
				strings.ReplaceAll(a.Name, "\"", "'"),
				a.Mailbox, a.Host)
		}
	} else {
		return fmt.Sprintf("<%s@%s>", a.Mailbox, a.Host)
	}
}

// FormatAddresses formats a list of addresses, separating each by a comma
func FormatAddresses(addrs []*Address) string {
	val := bytes.Buffer{}
	for i, addr := range addrs {
		val.WriteString(addr.Format())
		if i != len(addrs)-1 {
			val.WriteString(", ")
		}
	}
	return val.String()
}
