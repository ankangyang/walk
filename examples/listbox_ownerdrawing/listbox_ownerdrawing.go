// Copyright 2019 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lxn/win"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var mw *walk.MainWindow
	var lb *walk.ListBox
	var items []logEntry
	for i := 10000; i > 0; i-- {
		items = append(items, logEntry{time.Now().Add(-time.Second * time.Duration(i)), fmt.Sprintf("Some stuff just happend. %s", strings.Repeat("blah ", i%100))})
	}
	model := &logModel{items: items}
	styler := &Styler{
		lb:                  &lb,
		model:               model,
		dpi2StampSize:       make(map[int]walk.Size),
		widthDPI2WsPerLine:  make(map[widthDPI]int),
		textWidthDPI2Height: make(map[textWidthDPI]int),
	}

	if err := (MainWindow{
		AssignTo: &mw,
		Title: "Walk ListBox Owner Drawing Example",
		MinSize:  Size{600, 400},
		Font:     Font{Family: "Segoe UI", PointSize: 9},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				DoubleBuffering: true,
				Layout:          VBox{},
				Children: []Widget{
					ListBox{
						AssignTo:   &lb,
						MultiSelection: true,
						Model:      model,
						ItemStyler: styler,
					},
				},
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	mw.Show()
	mw.Run()
}

type logModel struct {
	walk.ReflectListModelBase
	items []logEntry
}

func (m *logModel) Items() interface{} {
	return m.items
}

type logEntry struct {
	timestamp time.Time
	message   string
}

type widthDPI struct {
	width int
	dpi   int
}

type textWidthDPI struct {
	text  string
	width int
	dpi   int
}

type Styler struct {
	lb                  **walk.ListBox
	canvas *walk.Canvas
	model               *logModel
	dpi2StampSize       map[int]walk.Size
	widthDPI2WsPerLine  map[widthDPI]int
	textWidthDPI2Height map[textWidthDPI]int
}

func (s *Styler) ItemHeightDependsOnWidth() bool {
	return true
}

func (s *Styler) DefaultItemHeight() int {
	return s.StampSize().Height+8
}

func (s *Styler) ItemHeight(index, width int) int {
	dpi := (*s.lb).DPI()

	msg := s.model.items[index].message

	twd := textWidthDPI{msg, width, dpi}

	if height, ok := s.textWidthDPI2Height[twd]; ok {
		return height+8
	}

	canvas, err := s.Canvas()
	if err != nil {
		return 0
	}

	stampSize := s.StampSize()

	wd := widthDPI{width, dpi}
	wsPerLine, ok := s.widthDPI2WsPerLine[wd]
	if !ok {
		bounds, _, err := canvas.MeasureText("W", (*s.lb).Font(), walk.Rectangle{Width: 9999999}, walk.TextCalcRect)
		if err != nil {
			return 0
		}
		wsPerLine = (width - stampSize.Width - 3 - 7 - 3) / bounds.Width
		s.widthDPI2WsPerLine[wd] = wsPerLine
	}

	if len(msg) <= wsPerLine {
		s.textWidthDPI2Height[twd] = stampSize.Height
		return stampSize.Height + 8
	}

	bounds, _, err := canvas.MeasureText(msg, (*s.lb).Font(), walk.Rectangle{Width: width - stampSize.Width - 3 - 7 - 3, Height: 255}, walk.TextEditControl|walk.TextWordbreak|walk.TextEndEllipsis)
	if err != nil {
		return 0
	}

	s.textWidthDPI2Height[twd] = bounds.Height

	return bounds.Height + 8
}

func (s *Styler) StyleItem(style *walk.ListItemStyle) {
	if canvas := style.Canvas(); canvas != nil {
		if style.Index()%2 == 1 && style.BackgroundColor == walk.Color(win.GetSysColor(win.COLOR_WINDOW)) {
			style.BackgroundColor = walk.Color(win.GetSysColor(win.COLOR_BTNFACE))
			if err := style.DrawBackground(); err != nil {
				return
			}
		}

		pen, err := walk.NewCosmeticPen(walk.PenSolid, walk.RGB(0, 0, 0))
		if err != nil {
			return
		}
		defer pen.Dispose()

		b := style.Bounds()
		b.X += 3
		b.Y += 4

		canvas.DrawText(s.model.items[style.Index()].timestamp.Format(time.StampMilli), (*s.lb).Font(), style.TextColor, b, walk.TextEditControl|walk.TextWordbreak)

		stampSize := s.StampSize()

		x := b.X + stampSize.Width + 3
		canvas.DrawLine(pen, walk.Point{x, b.Y - 4}, walk.Point{x, b.Y - 4 + b.Height})

		b.X += stampSize.Width + 7
		b.Width -= stampSize.Width + 3 + 7 + 3

		canvas.DrawText(s.model.items[style.Index()].message, (*s.lb).Font(), style.TextColor, b, walk.TextEditControl|walk.TextWordbreak|walk.TextEndEllipsis)
	}
}

func (s *Styler) StampSize()  walk.Size {
	dpi := (*s.lb).DPI()

	stampSize, ok := s.dpi2StampSize[dpi]
	if !ok {
		canvas, err := s.Canvas()
		if err != nil {
			return walk.Size{}
		}

		bounds, _, err := canvas.MeasureText("Jan _2 20:04:05.000", (*s.lb).Font(), walk.Rectangle{Width: 9999999}, walk.TextCalcRect)
		if err != nil {
			return  walk.Size{}
		}

		stampSize = bounds.Size()
		s.dpi2StampSize[dpi] = stampSize
	}

	return stampSize
}

func (s *Styler) Canvas() (*walk.Canvas, error) {
	if s.canvas != nil {
		return s.canvas, nil
	}
	
	canvas, err := (*s.lb).CreateCanvas()
	if err != nil {
		return nil, err
	}
	s.canvas = canvas
	(*s.lb).AddDisposable(canvas)

	return canvas, nil
}
