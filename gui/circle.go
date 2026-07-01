package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// MinSizeCircle wraps canvas.Circle to allow setting custom minimum sizes.
type MinSizeCircle struct {
	canvas.Circle
	minSize fyne.Size
}

// NewMinSizeCircle creates a new MinSizeCircle with the given color.
func NewMinSizeCircle(c color.Color) *MinSizeCircle {
	circle := &MinSizeCircle{}
	circle.FillColor = c
	return circle
}

// MinSize overrides the standard MinSize implementation.
func (m *MinSizeCircle) MinSize() fyne.Size {
	if m.minSize.Width == 0 && m.minSize.Height == 0 {
		return fyne.NewSize(1, 1)
	}
	return m.minSize
}

// SetMinSize updates the minimum size of this circle.
func (m *MinSizeCircle) SetMinSize(s fyne.Size) {
	m.minSize = s
}
