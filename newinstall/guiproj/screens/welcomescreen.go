package screens

import (
	"net/url"
	"newinstall/guiproj/data"

	"fyne.io/fyne/canvas"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func WelcomeScreen() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneScene)
	logo.SetMinSize(fyne.NewSize(228, 167))

	link, err := url.Parse("https://www.inovatian.com/")
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return widget.NewVBox(
		widget.NewLabelWithStyle("Welcome to Inovatian App", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewHBox(layout.NewSpacer(), logo, layout.NewSpacer()),
		widget.NewHyperlinkWithStyle("inovatian", link, fyne.TextAlignCenter, fyne.TextStyle{}),
		layout.NewSpacer(),

		widget.NewGroup("",
			fyne.NewContainerWithLayout(layout.NewGridLayout(2),

				widget.NewButton("Cancel", func() {
					// a.Settings().SetTheme(theme.LightTheme())
					App.Quit()
				}),
				widget.NewButton("Next", func() {
					// a.Settings().SetTheme(theme.DarkTheme())
					// set content
					Wind.SetContent(BuildContent())

				}),
			),
		),
	)
}
