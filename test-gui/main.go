package main

import (
	"time"

	"fyne.io/fyne"

	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func showTime(clock *widget.Label) {
	formatted := time.Now().Format("03:04:05")
	clock.SetText(formatted)
}
func main() {
	app := app.New()
	entry := widget.NewEntry()
	entry.SetText("Path")
	lbl := widget.NewLabel("Hello Fyne!")
	w := app.NewWindow("Hello")
	vbox := widget.NewVBox(
		lbl,
		entry,
		widget.NewButton("Quit", func() {
			// app.Quit()
			// w2 := app.NewWindow("Second")
			// w.Close()
			// w2.CenterOnScreen()
			// w2.SetContent(buildContent())
			// // w2.SetFullScreen(true)
			// w2.Resize(fyne.NewSize(240, 180))
			// w2.ShowAndRun()
			w.SetContent(buildContent())
		}),
		widget.NewButton("Quit", func() {
			app.Quit()
		}),
	)

	w.SetContent(vbox)
	w.CenterOnScreen()
	//w.SetFullScreen(true)
	w.Resize(fyne.NewSize(240, 180))
	w.ShowAndRun()
}
func buildContent() *fyne.Container {
	lbTop := widget.NewLabel("@top")
	lbTop.Alignment = fyne.TextAlignCenter

	lbBottom := widget.NewLabel("@bottom")
	lbBottom.Alignment = fyne.TextAlignCenter

	lbLeft := widget.NewLabel("@left")
	lbLeft.Alignment = fyne.TextAlignCenter

	lbRight := widget.NewLabel("@right")
	lbRight.Alignment = fyne.TextAlignCenter

	lbCenter := widget.NewLabelWithStyle("@center",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	return fyne.NewContainerWithLayout(
		layout.NewBorderLayout(lbTop, lbBottom, lbLeft, lbRight),
		lbTop, lbBottom, lbLeft, lbRight,
		lbCenter,
	)
}
