package gui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// buildAboutScreen renders descriptive application details and credits.
func (a *CodeForgeApp) buildAboutScreen() fyne.CanvasObject {
	logo := LoadLogo()
	img := canvas.NewImageFromResource(logo)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(120, 120))

	title := canvas.NewText("CodeForge", a.FyneApp.Settings().Theme().Color("primary", 0))
	title.TextSize = 22
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	tagline := widget.NewLabel("CI/CD - v1.0.0")
	tagline.Alignment = fyne.TextAlignCenter

	credit := widget.NewLabel("Built with ♥ by KhajumSanjog")
	credit.Alignment = fyne.TextAlignCenter

	desc := widget.NewLabel("A CI/CD daemon with its own\nconfiguration scripting language")
	desc.Alignment = fyne.TextAlignCenter

	techInfo := widget.NewLabel("Go 1.22  ·  Fyne v2  ·  KZM Language")
	techInfo.Alignment = fyne.TextAlignCenter

	ghURL, _ := url.Parse("https://github.com/KhajumSanjog/CodeForge")
	docURL, _ := url.Parse("https://github.com/KhajumSanjog/CodeForge/wiki")

	ghLink := widget.NewHyperlink("View on GitHub", ghURL)
	docLink := widget.NewHyperlink("Documentation", docURL)

	links := container.NewHBox(layout.NewSpacer(), ghLink, widget.NewLabel("·"), docLink, layout.NewSpacer())

	copyright := widget.NewLabel("© 2025 KhajumSanjog")
	copyright.Alignment = fyne.TextAlignCenter

	aboutBox := container.NewVBox(
		widget.NewLabel(""),
		container.NewCenter(img),
		title,
		tagline,
		credit,
		desc,
		widget.NewSeparator(),
		techInfo,
		links,
		widget.NewSeparator(),
		copyright,
	)

	return container.NewCenter(aboutBox)
}
