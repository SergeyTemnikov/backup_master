package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
)

type VariantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (t *VariantTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return t.base.Color(n, t.variant)
}

func (t *VariantTheme) Font(s fyne.TextStyle) fyne.Resource {
	return t.base.Font(s)
}

func (t *VariantTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(n)
}

func (t *VariantTheme) Size(n fyne.ThemeSizeName) float32 {
	return t.base.Size(n)
}
