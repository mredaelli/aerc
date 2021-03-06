//+build notmuch

package notmuch

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"git.sr.ht/~sircmpwn/aerc/models"
	"git.sr.ht/~sircmpwn/aerc/worker/lib"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	notmuch "github.com/zenhack/go.notmuch"
)

type Message struct {
	uid uint32
	key string
	msg *notmuch.Message
}

// NewReader reads a message into memory and returns an io.Reader for it.
func (m Message) NewReader() (io.Reader, error) {
	f, err := os.Open(m.msg.Filename())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// MessageInfo populates a models.MessageInfo struct for the message.
func (m Message) MessageInfo() (*models.MessageInfo, error) {
	return lib.MessageInfo(m)
}

// NewBodyPartReader creates a new io.Reader for the requested body part(s) of
// the message.
func (m Message) NewBodyPartReader(requestedParts []int) (io.Reader, error) {
	f, err := os.Open(m.msg.Filename())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	msg, err := message.Read(f)
	if err != nil {
		return nil, fmt.Errorf("could not read message: %v", err)
	}
	return lib.FetchEntityPartReader(msg, requestedParts)
}

// MarkRead either adds or removes the maildir.FlagSeen flag from the message.
func (m Message) MarkRead(seen bool) error {
	haveUnread := false
	for _, t := range m.tags() {
		if t == "unread" {
			haveUnread = true
			break
		}
	}
	if (haveUnread && !seen) || (!haveUnread && seen) {
		// we already have the desired state
		return nil
	}

	if haveUnread {
		err := m.msg.RemoveTag("unread")
		if err != nil {
			return err
		}
		return nil
	}

	err := m.msg.AddTag("unread")
	if err != nil {
		return err
	}
	return nil
}

// tags returns the notmuch tags of a message
func (m Message) tags() []string {
	ts := m.msg.Tags()
	var tags []string
	var tag *notmuch.Tag
	for ts.Next(&tag) {
		tags = append(tags, tag.Value)
	}
	return tags
}

func (m Message) ModelFlags() ([]models.Flag, error) {
	var flags []models.Flag
	seen := true

	for _, tag := range m.tags() {
		switch tag {
		case "replied":
			flags = append(flags, models.AnsweredFlag)
		case "flagged":
			flags = append(flags, models.FlaggedFlag)
		case "unread":
			seen = false
		default:
			continue
		}
	}
	if seen {
		flags = append(flags, models.SeenFlag)
	}
	return flags, nil
}

func (m Message) UID() uint32 {
	return m.uid
}
