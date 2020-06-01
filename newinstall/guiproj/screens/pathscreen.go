package screens

import (
	"fmt"
	"newinstall/guiproj/utils"
	"os"
	"os/user"
	"path/filepath"

	"fyne.io/fyne/app"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

// App app instance
var App = app.New()

// Wind window instance
var Wind = App.NewWindow("ino-storage")

// Defaultpath where store chunks
var Defaultpath string

//BuildContent build first screen content
func BuildContent() *fyne.Container {
	lbpath := widget.NewLabel("path")
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}
	Defaultpath := filepath.Join(usr.HomeDir, "ino-storage")
	lbpathcontent := widget.NewLabel(Defaultpath)

	ret := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	//	ret.Layout.Layout()
	ret.AddObject(lbpath)
	ret.AddObject(lbpathcontent)
	ret.AddObject(layout.NewSpacer())
	row := widget.NewGroup("",
		fyne.NewContainerWithLayout(layout.NewGridLayout(2),

			widget.NewButton("Back", func() {
				Wind.SetContent(WelcomeScreen())

			}),
			widget.NewButton("Next", func() {
				os.Mkdir(Defaultpath, 0700)
				utils.Defaultpath = Defaultpath
				Wind.SetContent(ResourcesScreen())
			}),
		),
	)

	ret.AddObject(row)
	return ret
}
