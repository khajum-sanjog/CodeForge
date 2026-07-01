package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CodeForgeTheme implements the fyne.Theme interface to define customized brand colors.
type CodeForgeTheme struct{}

// Color returns the brand-specific color palette for widgets.
func (t CodeForgeTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x53, G: 0x4A, B: 0xB7, A: 0xFF} // #534AB7 (purple)
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x1A, G: 0x1A, B: 0x2E, A: 0xFF} // #1a1a2e (dark background)
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xF5, G: 0xF4, B: 0xFF, A: 0xFF} // #f5f4ff (light foreground)
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x0F, G: 0x6E, B: 0x56, A: 0xFF} // #0F6E56 (green success)
	case theme.ColorNameError:
		return color.NRGBA{R: 0xD8, G: 0x5A, B: 0x30, A: 0xFF} // #D85A30 (red error)
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xBA, G: 0x75, B: 0x17, A: 0xFF} // #BA7517 (amber warning)
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x24, G: 0x24, B: 0x3B, A: 0xFF}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x53, G: 0x4A, B: 0xB7, A: 0xFF}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0x66}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

// Font delegates typography loading to the Fyne default fonts.
func (t CodeForgeTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon delegates system icon loading to the Fyne default icons.
func (t CodeForgeTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns standard layout dimension guidelines.
func (t CodeForgeTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
