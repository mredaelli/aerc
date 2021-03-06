package maildir

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/emersion/go-maildir"
	"github.com/fsnotify/fsnotify"

	"git.sr.ht/~sircmpwn/aerc/models"
	"git.sr.ht/~sircmpwn/aerc/worker/handlers"
	"git.sr.ht/~sircmpwn/aerc/worker/types"
)

func init() {
	handlers.RegisterWorkerFactory("maildir", NewWorker)
}

var errUnsupported = fmt.Errorf("unsupported command")

// A Worker handles interfacing between aerc's UI and a group of maildirs.
type Worker struct {
	c        *Container
	selected *maildir.Dir
	worker   *types.Worker
	watcher  *fsnotify.Watcher
}

// NewWorker creates a new maildir worker with the provided worker.
func NewWorker(worker *types.Worker) (types.Backend, error) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("could not create file system watcher: %v", err)
	}
	return &Worker{worker: worker, watcher: watch}, nil
}

// Run starts the worker's message handling loop.
func (w *Worker) Run() {
	for {
		select {
		case action := <-w.worker.Actions:
			w.handleAction(action)
		case ev := <-w.watcher.Events:
			w.handleFSEvent(ev)
		}
	}
}

func (w *Worker) handleAction(action types.WorkerMessage) {
	msg := w.worker.ProcessAction(action)
	if err := w.handleMessage(msg); err == errUnsupported {
		w.worker.PostMessage(&types.Unsupported{
			Message: types.RespondTo(msg),
		}, nil)
	} else if err != nil {
		w.worker.PostMessage(&types.Error{
			Message: types.RespondTo(msg),
			Error:   err,
		}, nil)
	} else {
		w.done(msg)
	}
}

func (w *Worker) handleFSEvent(ev fsnotify.Event) {
	// we only care about files being created
	if ev.Op != fsnotify.Create {
		return
	}
	// if there's not a selected directory to rescan, ignore
	if w.selected == nil {
		return
	}
	_, err := w.selected.Unseen()
	if err != nil {
		w.worker.Logger.Printf("could not move new to cur : %v", err)
		return
	}
	uids, err := w.c.UIDs(*w.selected)
	if err != nil {
		w.worker.Logger.Printf("could not scan UIDs: %v", err)
		return
	}
	w.worker.PostMessage(&types.DirectoryContents{
		Uids: uids,
	}, nil)
}

func (w *Worker) done(msg types.WorkerMessage) {
	w.worker.PostMessage(&types.Done{types.RespondTo(msg)}, nil)
}

func (w *Worker) err(msg types.WorkerMessage, err error) {
	w.worker.PostMessage(&types.Error{
		Message: types.RespondTo(msg),
		Error:   err,
	}, nil)
}

func (w *Worker) handleMessage(msg types.WorkerMessage) error {
	switch msg := msg.(type) {
	case *types.Unsupported:
		// No-op
	case *types.Configure:
		return w.handleConfigure(msg)
	case *types.Connect:
		return w.handleConnect(msg)
	case *types.ListDirectories:
		return w.handleListDirectories(msg)
	case *types.OpenDirectory:
		return w.handleOpenDirectory(msg)
	case *types.FetchDirectoryContents:
		return w.handleFetchDirectoryContents(msg)
	case *types.CreateDirectory:
		return w.handleCreateDirectory(msg)
	case *types.FetchMessageHeaders:
		return w.handleFetchMessageHeaders(msg)
	case *types.FetchMessageBodyPart:
		return w.handleFetchMessageBodyPart(msg)
	case *types.FetchFullMessages:
		return w.handleFetchFullMessages(msg)
	case *types.DeleteMessages:
		return w.handleDeleteMessages(msg)
	case *types.ReadMessages:
		return w.handleReadMessages(msg)
	case *types.CopyMessages:
		return w.handleCopyMessages(msg)
	case *types.AppendMessage:
		return w.handleAppendMessage(msg)
	case *types.SearchDirectory:
		return w.handleSearchDirectory(msg)
	}
	return errUnsupported
}

func (w *Worker) handleConfigure(msg *types.Configure) error {
	u, err := url.Parse(msg.Config.Source)
	if err != nil {
		w.worker.Logger.Printf("error configuring maildir worker: %v", err)
		return err
	}
	dir := u.Path
	if u.Host == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not resolve home directory: %v", err)
		}
		dir = filepath.Join(home, u.Path)
	}
	w.c = NewContainer(dir, w.worker.Logger)
	w.worker.Logger.Printf("configured base maildir: %s", dir)
	return nil
}

func (w *Worker) handleConnect(msg *types.Connect) error {
	return nil
}

func (w *Worker) handleListDirectories(msg *types.ListDirectories) error {
	dirs, err := w.c.ListFolders()
	if err != nil {
		w.worker.Logger.Printf("error listing directories: %v", err)
		return err
	}
	for _, name := range dirs {
		w.worker.PostMessage(&types.Directory{
			Message: types.RespondTo(msg),
			Dir: &models.Directory{
				Name:       name,
				Attributes: []string{},
			},
		}, nil)
	}
	return nil
}

func (w *Worker) handleOpenDirectory(msg *types.OpenDirectory) error {
	w.worker.Logger.Printf("opening %s", msg.Directory)

	// remove existing watch path
	if w.selected != nil {
		prevDir := filepath.Join(string(*w.selected), "new")
		if err := w.watcher.Remove(prevDir); err != nil {
			return fmt.Errorf("could not unwatch previous directory: %v", err)
		}
	}

	// open the directory
	dir, err := w.c.OpenDirectory(msg.Directory)
	if err != nil {
		return err
	}
	w.selected = &dir

	// add watch path
	newDir := filepath.Join(string(*w.selected), "new")
	if err := w.watcher.Add(newDir); err != nil {
		return fmt.Errorf("could not add watch to directory: %v", err)
	}

	if err := dir.Clean(); err != nil {
		return fmt.Errorf("could not clean directory: %v", err)
	}

	// TODO: why does this need to be sent twice??
	info := &types.DirectoryInfo{
		Info: &models.DirectoryInfo{
			Name:     msg.Directory,
			Flags:    []string{},
			ReadOnly: false,
			// total messages
			Exists: 0,
			// new messages since mailbox was last opened
			Recent: 0,
			// total unread
			Unseen: 0,
		},
	}
	w.worker.PostMessage(info, nil)
	w.worker.PostMessage(info, nil)
	return nil
}

func (w *Worker) handleFetchDirectoryContents(
	msg *types.FetchDirectoryContents) error {
	uids, err := w.c.UIDs(*w.selected)
	if err != nil {
		w.worker.Logger.Printf("error scanning uids: %v", err)
		return err
	}
	w.worker.PostMessage(&types.DirectoryContents{
		Message: types.RespondTo(msg),
		Uids:    uids,
	}, nil)
	return nil
}

func (w *Worker) handleCreateDirectory(msg *types.CreateDirectory) error {
	dir := w.c.Dir(msg.Directory)
	if err := dir.Create(); err != nil {
		w.worker.Logger.Printf("could not create directory %s: %v",
			msg.Directory, err)
		return err
	}
	return nil
}

func (w *Worker) handleFetchMessageHeaders(
	msg *types.FetchMessageHeaders) error {
	for _, uid := range msg.Uids {
		m, err := w.c.Message(*w.selected, uid)
		if err != nil {
			w.worker.Logger.Printf("could not get message: %v", err)
			w.err(msg, err)
			continue
		}
		info, err := m.MessageInfo()
		if err != nil {
			w.worker.Logger.Printf("could not get message info: %v", err)
			w.err(msg, err)
			continue
		}
		w.worker.PostMessage(&types.MessageInfo{
			Message: types.RespondTo(msg),
			Info:    info,
		}, nil)
	}
	return nil
}

func (w *Worker) handleFetchMessageBodyPart(
	msg *types.FetchMessageBodyPart) error {

	// get reader
	m, err := w.c.Message(*w.selected, msg.Uid)
	if err != nil {
		w.worker.Logger.Printf("could not get message %d: %v", msg.Uid, err)
		return err
	}
	r, err := m.NewBodyPartReader(msg.Part)
	if err != nil {
		w.worker.Logger.Printf(
			"could not get body part reader for message=%d, parts=%#v: %v",
			msg.Uid, msg.Part, err)
		return err
	}
	w.worker.PostMessage(&types.MessageBodyPart{
		Message: types.RespondTo(msg),
		Part: &models.MessageBodyPart{
			Reader: r,
			Uid:    msg.Uid,
		},
	}, nil)

	// mark message as read
	if err := m.MarkRead(true); err != nil {
		w.worker.Logger.Printf("could not mark message as read: %v", err)
		return err
	}

	// send updated flags to ui
	info, err := m.MessageInfo()
	if err != nil {
		w.worker.Logger.Printf("could not fetch message info: %v", err)
		return err
	}
	w.worker.PostMessage(&types.MessageInfo{
		Message: types.RespondTo(msg),
		Info:    info,
	}, nil)

	return nil
}

func (w *Worker) handleFetchFullMessages(msg *types.FetchFullMessages) error {
	for _, uid := range msg.Uids {
		m, err := w.c.Message(*w.selected, uid)
		if err != nil {
			w.worker.Logger.Printf("could not get message %d: %v", uid, err)
			return err
		}
		r, err := m.NewReader()
		if err != nil {
			w.worker.Logger.Printf("could not get message reader: %v", err)
			return err
		}
		w.worker.PostMessage(&types.FullMessage{
			Message: types.RespondTo(msg),
			Content: &models.FullMessage{
				Uid:    uid,
				Reader: r,
			},
		}, nil)
	}
	return nil
}

func (w *Worker) handleDeleteMessages(msg *types.DeleteMessages) error {
	deleted, err := w.c.DeleteAll(*w.selected, msg.Uids)
	if len(deleted) > 0 {
		w.worker.PostMessage(&types.MessagesDeleted{
			Message: types.RespondTo(msg),
			Uids:    deleted,
		}, nil)
	}
	if err != nil {
		w.worker.Logger.Printf("error removing some messages: %v", err)
		return err
	}
	return nil
}

func (w *Worker) handleReadMessages(msg *types.ReadMessages) error {
	for _, uid := range msg.Uids {
		m, err := w.c.Message(*w.selected, uid)
		if err != nil {
			w.worker.Logger.Printf("could not get message: %v", err)
			w.err(msg, err)
			continue
		}
		if err := m.MarkRead(msg.Read); err != nil {
			w.worker.Logger.Printf("could not mark message as read: %v", err)
			w.err(msg, err)
			continue
		}
		info, err := m.MessageInfo()
		if err != nil {
			w.worker.Logger.Printf("could not get message info: %v", err)
			w.err(msg, err)
			continue
		}
		w.worker.PostMessage(&types.MessageInfo{
			Message: types.RespondTo(msg),
			Info:    info,
		}, nil)
	}
	return nil
}

func (w *Worker) handleCopyMessages(msg *types.CopyMessages) error {
	dest := w.c.Dir(msg.Destination)
	err := w.c.CopyAll(dest, *w.selected, msg.Uids)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) handleAppendMessage(msg *types.AppendMessage) error {
	dest := w.c.Dir(msg.Destination)
	delivery, err := dest.NewDelivery()
	if err != nil {
		w.worker.Logger.Printf("could not deliver message to %s: %v",
			msg.Destination, err)
		return err
	}
	defer delivery.Close()
	if _, err := io.Copy(delivery, msg.Reader); err != nil {
		w.worker.Logger.Printf("could not write message to destination: %v", err)
		return err
	}
	return nil
}

func (w *Worker) handleSearchDirectory(msg *types.SearchDirectory) error {
	return errUnsupported
}
