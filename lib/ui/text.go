package ui

import (
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

const (
	TEXT_LEFT   = iota
	TEXT_CENTER = iota
	TEXT_RIGHT  = iota
)

type Text struct {
	Invalidatable
	text     string
	strategy uint
	fg       tcell.Color
	bg       tcell.Color
	bold     bool
	reverse  bool
}

func NewText(text string) *Text {
	return &Text{
		bg:   tcell.ColorDefault,
		fg:   tcell.ColorDefault,
		text: text,
	}
}

func (t *Text) Text(text string) *Text {
	t.text = text
	t.Invalidate()
	return t
}

func (t *Text) Strategy(strategy uint) *Text {
	t.strategy = strategy
	t.Invalidate()
	return t
}

func (t *Text) Bold(bold bool) *Text {
	t.bold = bold
	t.Invalidate()
	return t
}

func (t *Text) Color(fg tcell.Color, bg tcell.Color) *Text {
	t.fg = fg
	t.bg = bg
	t.Invalidate()
	return t
}

func (t *Text) Reverse(reverse bool) *Text {
	t.reverse = reverse
	t.Invalidate()
	return t
}

func (t *Text) Draw(ctx *Context) {
	size := runewidth.StringWidth(t.text)
	x := 0
	if t.strategy == TEXT_CENTER {
		x = (ctx.Width() - size) / 2
	}
	if t.strategy == TEXT_RIGHT {
		x = ctx.Width() - size
	}
	style := tcell.StyleDefault.Background(t.bg).Foreground(t.fg)
	if t.bold {
		style = style.Bold(true)
	}
	if t.reverse {
		style = style.Reverse(true)
	}
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', style)
	ctx.Printf(x, 0, style, "%s", t.text)
}

func (t *Text) Invalidate() {
	t.DoInvalidate(t)
}
