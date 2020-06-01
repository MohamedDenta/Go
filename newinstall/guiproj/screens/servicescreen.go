package screens

import (
	"net/url"
	"newinstall/guiproj/data"
	"newinstall/guiproj/utils"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var execstartpath string = "/bin/inostorage"
var autoservice string = `[Unit]
Description=Sleep service
ConditionPathExists=` + execstartpath + `
After=network.target
[Service]
Type=simple
Restart=on-failure
RestartSec=30
startLimitIntervalSec=60
# WorkingDirectory=/home/moha/go/src/service-demo
ExecStart= ` + execstartpath + ` 
# # make sure log directory exists and owned by syslog
# PermissionsStartOnly=true
# ExecStartPre=/bin/mkdir -p /var/log/sleepservice
# ExecStartPre=/bin/chown syslog:adm /var/log/sleepservice
# ExecStartPre=/bin/chmod 755 /var/log/sleepservice
# StandardOutput=syslog
# StandardError=syslog
# SyslogIdentifier=sleepservice
 
[Install]
WantedBy=multi-user.target`

func ServiceScreen() fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneScene)
	logo.SetMinSize(fyne.NewSize(228, 167))

	link, err := url.Parse("https://www.inovatian.com/")
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}
	btncancel := widget.NewButton("Cancel", func() {
		// a.Settings().SetTheme(theme.LightTheme())
		App.Quit()
	})
	btnfinish := widget.NewButton("Finish", func() {
		// a.Settings().SetTheme(theme.DarkTheme())
		// set content
		utils.WriteService(&autoservice)
		Wind.Hide()
		time.Sleep(time.Second * 1)
		Wind.SetContent(RegisterScreen())

		Wind.Show()
	})
	btnfinish.Hide()
	progbar := widget.NewProgressBar()
	go func() {
		time.Sleep(100 * time.Millisecond)
		for progbar.Value < progbar.Max {
			time.Sleep(100 * time.Millisecond)
			progbar.SetValue(float64(progbar.Value + 0.01))
		}
		btnfinish.Show()
	}()

	return widget.NewVBox(
		widget.NewLabelWithStyle("Welcome to Inovatian App", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		widget.NewHBox(layout.NewSpacer(), logo, layout.NewSpacer()),
		widget.NewHyperlinkWithStyle("inovatian", link, fyne.TextAlignCenter, fyne.TextStyle{}),
		progbar,
		layout.NewSpacer(),

		widget.NewGroup("",
			fyne.NewContainerWithLayout(layout.NewGridLayout(2),
				btncancel,
				btnfinish,
			),
		),
	)

}
