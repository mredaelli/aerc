package ui

import (
	tb "github.com/nsf/termbox-go"

	"git.sr.ht/~sircmpwn/aerc2/config"
)

const (
	Valid             = 0
	InvalidateTabList = 1 << iota
	InvalidateTabView
	InvalidateSidebar
	InvalidateStatusBar
)

const (
	InvalidateAll = InvalidateTabList |
		InvalidateTabView |
		InvalidateSidebar |
		InvalidateStatusBar
)

type Geometry struct {
	Row    int
	Col    int
	Width  int
	Height int
}

type AercTab interface {
	Name() string
	Render(at Geometry)
	SetParent(parent *UIState)
}

type UIState struct {
	Config       *config.AercConfig
	Exit         bool
	InvalidPanes uint

	Panes struct {
		TabList   Geometry
		TabView   Geometry
		Sidebar   Geometry
		StatusBar Geometry
	}

	Tabs        []AercTab
	SelectedTab int

	Prompt struct {
		Prompt *string
		Text   *string
		Index  int
		Scroll int
	}

	tbEvents chan tb.Event
}