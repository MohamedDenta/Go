package screens

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"newinstall/guiproj/deviceinfo"
	"newinstall/guiproj/net"
	"newinstall/guiproj/utils"
	"path/filepath"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

//BuildContent build first screen content
func RegisterScreen() *fyne.Container {
	Wind.SetTitle("Register node")
	mac := widget.NewEntry()
	mac.SetText(net.GetMacAddr())
	mac.Disable()

	location := widget.NewEntry()
	location.SetPlaceHolder("node location")
	location.Disable()

	alias := widget.NewEntry()
	alias.SetPlaceHolder("node alias")

	usrlocation := widget.NewEntry()
	usrlocation.SetPlaceHolder("user location")

	storagespace := widget.NewEntry()
	storagespace.SetPlaceHolder("storage space in GB")
	//storagespace.Text = "100"

	ownerpk := widget.NewEntry()
	ownerpk.SetPlaceHolder("owner public key")

	nodepk := widget.NewEntry()
	nodepk.SetPlaceHolder("node public key")

	nodepvk := widget.NewEntry()
	nodepvk.SetPlaceHolder("node private key")

	email := widget.NewEntry()
	email.SetPlaceHolder("email")

	ret := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	//	ret.Layout.Layout()
	ret.AddObject(mac)
	ret.AddObject(location)
	ret.AddObject(alias)
	ret.AddObject(usrlocation)
	ret.AddObject(storagespace)
	ret.AddObject(ownerpk)
	ret.AddObject(nodepk)
	ret.AddObject(nodepvk)
	ret.AddObject(email)
	ret.AddObject(layout.NewSpacer())
	row := widget.NewGroup("",
		fyne.NewContainerWithLayout(layout.NewGridLayout(2),

			widget.NewButton("Cancel", func() {
				App.Quit()

			}),
			widget.NewButton("Save", func() {
				// send data to our validators
				// reserve storage
				utils.ReserveStorage(5 << 10)
				changeinstallstatus()
				App.Quit()
			}),
		),
	)

	ret.AddObject(row)
	return ret
}

func changeinstallstatus() {
	var info deviceinfo.Info
	var fname string = "ino-config.conf"
	filename := filepath.Join(utils.Configpath, fname)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Panicf("failed reading data from file: %s", err)
	}
	json.Unmarshal(data, &info)

	info.Installed = true
	utils.Installed = true
	b, _ := json.Marshal(info)
	e := ioutil.WriteFile(filename, b, 0644)
	if e != nil {
		fmt.Println("^^^^^^^^^^^^^^^^^^^ ", e)
	}
}
