package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func Title(text string) fyne.CanvasObject {
	return widget.NewLabelWithStyle(
		text,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
}
