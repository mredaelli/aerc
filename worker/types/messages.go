package types

import (
	"io"
	"time"

	"github.com/emersion/go-imap"

	"git.sr.ht/~sircmpwn/aerc/config"
	"git.sr.ht/~sircmpwn/aerc/models"
)

type WorkerMessage interface {
	InResponseTo() WorkerMessage
	getId() int64
	setId(id int64)
}

type Message struct {
	inResponseTo WorkerMessage
	id           int64
}

func RespondTo(msg WorkerMessage) Message {
	return Message{
		inResponseTo: msg,
	}
}

func (m Message) InResponseTo() WorkerMessage {
	return m.inResponseTo
}

func (m Message) getId() int64 {
	return m.id
}

func (m *Message) setId(id int64) {
	m.id = id
}

// Meta-messages

type Done struct {
	Message
}

type Error struct {
	Message
	Error error
}

type Unsupported struct {
	Message
}

// Actions

type Configure struct {
	Message
	Config *config.AccountConfig
}

type Connect struct {
	Message
}

type Disconnect struct {
	Message
}

type ListDirectories struct {
	Message
}

type OpenDirectory struct {
	Message
	Directory string
}

type FetchDirectoryContents struct {
	Message
}

type SearchDirectory struct {
	Message
	Criteria *imap.SearchCriteria
}

type CreateDirectory struct {
	Message
	Directory string
	Quiet     bool
}

type FetchMessageHeaders struct {
	Message
	Uids []uint32
}

type FetchFullMessages struct {
	Message
	Uids []uint32
}

type FetchMessageBodyPart struct {
	Message
	Uid  uint32
	Part []int
}

type DeleteMessages struct {
	Message
	Uids []uint32
}

// Marks messages as read or unread
type ReadMessages struct {
	Message
	Read bool
	Uids []uint32
}

type CopyMessages struct {
	Message
	Destination string
	Uids        []uint32
}

type AppendMessage struct {
	Message
	Destination string
	Flags       []string
	Date        time.Time
	Reader      io.Reader
	Length      int
}

// Messages

type Directory struct {
	Message
	Dir *models.Directory
}

type DirectoryInfo struct {
	Message
	Info *models.DirectoryInfo
}

type DirectoryContents struct {
	Message
	Uids []uint32
}

type SearchResults struct {
	Message
	Uids []uint32
}

type MessageInfo struct {
	Message
	Info *models.MessageInfo
}

type FullMessage struct {
	Message
	Content *models.FullMessage
}

type MessageBodyPart struct {
	Message
	Part *models.MessageBodyPart
}

type MessagesDeleted struct {
	Message
	Uids []uint32
}
