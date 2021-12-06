package main

import (
	"fmt"
	"strings"
)

type SegmentColor struct {
	fg int
	bg int
}

func (s *SegmentColor) IsEmpty() bool {
	return s.fg == -1 && s.bg == -1
}

func fmtColorInternal(prefix int, code int) string {
	return fmt.Sprintf("\\[\\e[%d;5;%dm\\]", prefix, code)
}

func fmtColorFg(code int) string {
	return fmtColorInternal(38, code)
}

func fmtColorBg(code int) string {
	return fmtColorInternal(48, code)
}

func (s *SegmentColor) fmt() string {
	return fmtColorFg(s.fg) + fmtColorBg(s.bg)
}

type Segment struct {
	beginColor   SegmentColor
	currentColor SegmentColor
	headLines    []string
	text         string
}

func (s *Segment) safeText(texts ...string) string {
	text := strings.Join(texts, "")
	text = strings.Replace(text, "$", "\\\\\\$", -1)
	text = strings.Replace(text, "`", "\\\\\\`", -1)
	return text
}

func (s *Segment) AddHeadline(texts ...string) {
	s.headLines = append(s.headLines, s.safeText(texts...))
}

func (s *Segment) AddText(texts ...string) {
	s.text += s.safeText(texts...)
}

func (s *Segment) SetTextColor(clr SegmentColor) {
	if s.beginColor.IsEmpty() {
		s.beginColor = clr
	}
	s.currentColor = clr
	s.AddText(s.currentColor.fmt())
}

func (s *Segment) IsPrintable() bool {
	return len(s.text) > 0 && !s.currentColor.IsEmpty()
}

func makeSegment() Segment {
	return Segment{beginColor: SegmentColor{-1, -1}, currentColor: SegmentColor{-1, -1}, text: ""}
}

func resetColors() string {
	return "\\[\\e[0m\\]"
}
