package ui

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

type Tabs struct {
	Tabs       []*Tab
	TabStrip   *TabStrip
	TabContent *TabContent
	Selected   int
	history    []int

	onInvalidateStrip   func(d Drawable)
	onInvalidateContent func(d Drawable)
}

type Tab struct {
	Content Drawable
	Name    string
	onClose func() bool
	invalid bool
}

type TabStrip Tabs
type TabContent Tabs

func NewTabs() *Tabs {
	tabs := &Tabs{}
	tabs.TabStrip = (*TabStrip)(tabs)
	tabs.TabContent = (*TabContent)(tabs)
	tabs.history = []int{}
	return tabs
}

func (tabs *Tabs) Add(content Drawable, name string) *Tab {
	tab := &Tab{
		Content: content,
		Name:    name,
	}
	tabs.Tabs = append(tabs.Tabs, tab)
	tabs.TabStrip.Invalidate()
	content.OnInvalidate(tabs.invalidateChild)
	return tab
}

func (tabs *Tabs) invalidateChild(d Drawable) {
	if tabs.Selected >= len(tabs.Tabs) {
		return
	}

	if tabs.Tabs[tabs.Selected].Content == d {
		if tabs.onInvalidateContent != nil {
			tabs.onInvalidateContent(tabs.TabContent)
		}
	}
}

func (tabs *Tabs) Remove(content Drawable) {
	match := false
	for i, tab := range tabs.Tabs {
		if tab.Content == content {
			tabs.Tabs = append(tabs.Tabs[:i], tabs.Tabs[i+1:]...)
			tabs.removeHistory(i)
			match = true
			break
		}
	}
	if !match {
		return
	}
	index, ok := tabs.popHistory()
	if ok {
		tabs.Select(index)
	}
	tabs.TabStrip.Invalidate()
}

func (tabs *Tabs) Replace(contentSrc Drawable, contentTarget Drawable, name string) {
	replaceTab := &Tab{
		Content: contentTarget,
		Name:    name,
	}
	for i, tab := range tabs.Tabs {
		if tab.Content == contentSrc {
			tabs.Tabs[i] = replaceTab
			tabs.Select(i)
			break
		}
	}
	tabs.TabStrip.Invalidate()
	contentTarget.OnInvalidate(tabs.invalidateChild)
}

func (tabs *Tabs) Select(index int) {
	if index >= len(tabs.Tabs) {
		index = len(tabs.Tabs) - 1
	}

	if tabs.Selected != index {
		tabs.pushHistory(tabs.Selected)
		tabs.Selected = index
		tabs.TabStrip.Invalidate()
		tabs.TabContent.Invalidate()
	}
}

func (tabs *Tabs) SelectPrevious() bool {
	index, ok := tabs.popHistory()
	if !ok {
		return false
	}
	tabs.Select(index)
	return true
}

func (tabs *Tabs) pushHistory(index int) {
	tabs.history = append(tabs.history, index)
}

func (tabs *Tabs) popHistory() (int, bool) {
	lastIdx := len(tabs.history) - 1
	if lastIdx < 0 {
		return 0, false
	}
	item := tabs.history[lastIdx]
	tabs.history = tabs.history[:lastIdx]
	return item, true
}

func (tabs *Tabs) removeHistory(index int) {
	newHist := make([]int, 0, len(tabs.history))
	for i, item := range tabs.history {
		if item == index {
			continue
		}
		if item > index {
			item = item - 1
		}
		// dedup
		if i > 0 && len(newHist) > 0 && item == newHist[len(newHist)-1] {
			continue
		}
		newHist = append(newHist, item)
	}
	tabs.history = newHist
}

func (tabs *Tabs) MouseEvent(event tcell.Event) {
	switch event := event.(type) {
	case *tcell.EventMouse:
		if event.Buttons()&tcell.Button1 != 0 {
			x, y := event.Position()
			selectedTab, ok := tabs.TabStrip.Clicked(x, y)
			if ok {
				tabs.Select(selectedTab)
			}
		}
	}
}

// Registers a handler for the tab to react to a close event
func (tab *Tab) OnClose(fn func() bool) {
	tab.onClose = fn
}

func (tab *Tab) CanClose() bool {
	if tab.onClose != nil {
		return tab.onClose()
	} else {
		return true
	}
}

// TODO: Color repository
func (strip *TabStrip) Draw(ctx *Context) {
	x := 0
	for i, tab := range strip.Tabs {
		style := tcell.StyleDefault.Reverse(true)
		if strip.Selected == i {
			style = tcell.StyleDefault
		}
		tabWidth := 32
		if ctx.Width()-x < tabWidth {
			tabWidth = ctx.Width() - x - 2
		}
		trunc := runewidth.Truncate(tab.Name, tabWidth, "…")
		x += ctx.Printf(x, 0, style, " %s ", trunc)
		if x >= ctx.Width() {
			break
		}
	}
	style := tcell.StyleDefault.Reverse(true)
	ctx.Fill(x, 0, ctx.Width()-x, 1, ' ', style)
}

func (strip *TabStrip) Invalidate() {
	if strip.onInvalidateStrip != nil {
		strip.onInvalidateStrip(strip)
	}
}

func (strip *TabStrip) OnInvalidate(onInvalidate func(d Drawable)) {
	strip.onInvalidateStrip = onInvalidate
}

func (strip *TabStrip) Clicked(mouseX int, mouseY int) (int, bool) {
	x := 0
	if mouseY == 0 {
		for i, tab := range strip.Tabs {
			trunc := runewidth.Truncate(tab.Name, 32, "…")
			length := len(trunc) + 2
			if x <= mouseX && mouseX < x+length {
				return i, true
			}
			x += length
		}
	}
	return 0, false
}

func (content *TabContent) Children() []Drawable {
	children := make([]Drawable, len(content.Tabs))
	for i, tab := range content.Tabs {
		children[i] = tab.Content
	}
	return children
}

func (content *TabContent) Draw(ctx *Context) {
	if content.Selected >= len(content.Tabs) {
		width := ctx.Width()
		height := ctx.Height()
		ctx.Fill(0, 0, width, height, ' ', tcell.StyleDefault)
	}

	tab := content.Tabs[content.Selected]
	tab.Content.Draw(ctx)
}

func (content *TabContent) Invalidate() {
	if content.onInvalidateContent != nil {
		content.onInvalidateContent(content)
	}
	tab := content.Tabs[content.Selected]
	tab.Content.Invalidate()
}

func (content *TabContent) OnInvalidate(onInvalidate func(d Drawable)) {
	content.onInvalidateContent = onInvalidate
}
