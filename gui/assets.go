package gui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed logo.png
var logoBytes []byte

// LoadLogo returns the embedded CodeForge logo as a Fyne static resource.
func LoadLogo() fyne.Resource {
	return fyne.NewStaticResource("logo.png", logoBytes)
}
