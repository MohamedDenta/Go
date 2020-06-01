package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"newinstall/guiproj/deviceinfo"
	"newinstall/guiproj/screens"
	"newinstall/guiproj/utils"
	"path/filepath"

	"fyne.io/fyne"
	"fyne.io/fyne/theme"
)

var app = screens.App
var w = screens.Wind

func main() {
	readFile()
	app.Settings().SetTheme(theme.LightTheme())
	if !utils.Installed {
		//lbl := widget.NewLabel("Hello Fyne!")
		w.SetContent(screens.WelcomeScreen())

	} else {
		w.SetContent(screens.RegisterScreen())
	}
	//w.SetContent(vbox)
	w.CenterOnScreen()
	//w.SetFullScreen(true)
	w.Resize(fyne.NewSize(300, 300))
	w.ShowAndRun()
}
func readFile() {
	var info deviceinfo.Info
	var fname string = "ino-config.conf"
	filename := filepath.Join(utils.Configpath, fname)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	json.Unmarshal(data, &info)

	utils.Installed = info.Installed
}
