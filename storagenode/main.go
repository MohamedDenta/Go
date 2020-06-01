package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"./net"

	"./deviceinfo"

	"./utils"

	"github.com/andlabs/ui"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

var mainwin *ui.Window
var defaultpath string
var configpath string
var buildpath string
var buildpath2 string
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
ExecStart=` + execstartpath + ` --name=demo-service
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

func makeDataChoosersPage() ui.Control {
	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)
	vbox1 := ui.NewVerticalBox()
	vbox1.SetPadded(true)
	hbox.Append(vbox, true)
	hbox.Append(vbox1, true)

	entry2 := ui.NewEntry()
	entry2.SetText(defaultpath)
	entry2.SetReadOnly(true)
	vbox.Append(ui.NewLabel("installation path "), false)
	vbox.Append(entry2, false)

	button := ui.NewButton("Next")
	button.OnClicked(func(*ui.Button) {
		// New Window
		e := os.Mkdir(defaultpath, 0700)
		if e != nil {
			fmt.Println(e)
		}
		ReadResaurces()
		mainwin.Destroy()
	})
	vbox1.Append(button, false)
	button = ui.NewButton("Cancel")
	button.OnClicked(func(*ui.Button) {
		window2 := ui.NewWindow("", 500, 500, false)
		window2.SetMargined(true)

		button.OnClicked(func(*ui.Button) {
			// window2.Show()
			mainwin.Destroy()
			os.Exit(0)
		})

	})
	vbox1.Append(button, false)
	return hbox
}
func setupUI() {
	mainwin = ui.NewWindow("storage node demo", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})
	control := makeDataChoosersPage()
	mainwin.SetChild(control)
	mainwin.Show()
}

// ReadResaurces read res
func ReadResaurces() {
	var deviceinfobj deviceinfo.Info
	cpucores := deviceinfo.Getcpucores()
	vmStat, err := mem.VirtualMemory()
	dealwithErr(err)
	diskStat, err := disk.Usage("/")
	dealwithErr(err)
	cpuStat, err := cpu.Info()
	dealwithErr(err)
	_, err = cpu.Percent(0, true)
	dealwithErr(err)
	box1 := ui.NewVerticalBox()
	// ui
	hostStat, _ := host.Info()
	box1.Append(ui.NewLabel("cpu cores : "+strconv.Itoa(cpucores)), false)
	for i, cpu := range cpuStat {
		box1.Append(ui.NewLabel("Speed cpu : "+strconv.Itoa(i)+" "+strconv.Itoa(i)+strconv.FormatFloat(cpu.Mhz, 'f', 2, 64)+" MHz"), false)
		deviceinfobj.CpuInfo.CpuSpeed = append(deviceinfobj.CpuInfo.CpuSpeed, strconv.Itoa(i)+strconv.FormatFloat(cpu.Mhz, 'f', 2, 64))
	}
	box1.Append(ui.NewLabel("OS Type : "+hostStat.OS), false)
	deviceinfobj.OSVersion = hostStat.OS
	box1.Append(ui.NewLabel("OS Platform : "+hostStat.Platform), false)
	box1.Append(ui.NewLabel("total mem : "+strconv.FormatUint(vmStat.Total, 10)), false)
	deviceinfobj.MemoryInfo.TotalMemory = strconv.FormatUint(vmStat.Total, 10)
	box1.Append(ui.NewLabel("free mem : "+strconv.FormatUint(vmStat.Free, 10)), false)
	deviceinfobj.MemoryInfo.FreeMemory = strconv.FormatUint(vmStat.Free, 10)
	box1.Append(ui.NewLabel("used percent mem : "+strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64)+"%"), false)
	deviceinfobj.MemoryInfo.PercentageMemory = strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64)
	box1.Append(ui.NewLabel("Total disk space: "+strconv.FormatUint(diskStat.Total, 10)+" bytes"), false)
	deviceinfobj.DiskInfo.TotalDiskSpace = strconv.FormatUint(diskStat.Total, 10)
	box1.Append(ui.NewLabel("Used disk space: "+strconv.FormatUint(diskStat.Used, 10)+" bytes"), false)
	deviceinfobj.DiskInfo.UsedDiskSpace = strconv.FormatUint(diskStat.Used, 10)
	box1.Append(ui.NewLabel("Free disk space: "+strconv.FormatUint(diskStat.Free, 10)+" bytes"), false)
	deviceinfobj.DiskInfo.FreeDiskSpace = strconv.FormatUint(diskStat.Free, 10)
	box1.Append(ui.NewLabel("Percentage disk space usage: "+strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64)+"%"), false)
	deviceinfobj.DiskInfo.PercentageDiskSpace = strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64)
	networkStats, _ := net.IOCounters()
	for _, counter := range networkStats {
		// ifcounter.Name=="wlp2s0")
		// TODO: check interface name
		box1.Append(ui.NewLabel("name  "+counter.Name), false)
		box1.Append(ui.NewLabel("bytesSent "+utils.FormatBytes(counter.BytesSent, 2)), false)
		box1.Append(ui.NewLabel("bytesRecv "+utils.FormatBytes(counter.BytesRecv, 2)), false)
		box1.Append(ui.NewLabel("packetsSent "+utils.FormatBytes(counter.PacketsSent, 2)), false)
		box1.Append(ui.NewLabel("packetsRecv "+utils.FormatBytes(counter.PacketsRecv, 2)), false)

		deviceinfobj.NetworkInfo.Name = counter.Name
		deviceinfobj.NetworkInfo.ByteSent = utils.FormatBytes(counter.BytesSent, 2)
		deviceinfobj.NetworkInfo.ByteRcvd = utils.FormatBytes(counter.BytesRecv, 2)
		deviceinfobj.NetworkInfo.PacketSent = utils.FormatBytes(counter.PacketsSent, 2)
		deviceinfobj.NetworkInfo.PacketRcvd = utils.FormatBytes(counter.PacketsRecv, 2)
		break
	}
	ip, er := net.ExternalIP()
	if er != nil {
		fmt.Println(er)
	}
	mac := net.GetMacAddr()
	deviceinfobj.IP = ip
	deviceinfobj.MAC = mac
	box1.Append(ui.NewLabel("IP : "+ip), false)
	box1.Append(ui.NewLabel("MAC : "+mac), false)
	window2 := ui.NewWindow("Device info", 500, 500, false)
	window2.SetMargined(true)
	button := ui.NewButton("Next")
	button.OnClicked(func(*ui.Button) {
		deviceinfobj.StoragePath = defaultpath
		str, isfine := utils.WriteConfigFile(&deviceinfobj, &configpath)
		if !isfine {
			fmt.Println(str)
		}
		// new window
		screen3()
		window2.Destroy()
	})
	box1.Append(button, false)
	button = ui.NewButton("Cancel")
	box1.Append(button, false)
	// New window
	window2.SetChild(box1)
	window2.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		window2.Destroy()
		return true
	})

	window2.Show()

}
func screen3() {
	ip := ui.NewProgressBar()
	box1 := ui.NewVerticalBox()
	window2 := ui.NewWindow("Successful installation", 500, 500, false)
	window2.SetMargined(true)
	button := ui.NewButton("Finish")
	button.Disable()
	box1.Append(ip, false)
	box1.Append(button, false)
	window2.SetChild(box1)
	window2.Show()
	go func() {
		for index := -1; index <= 100; index++ {
			time.Sleep(time.Millisecond * 80)
			ip.SetValue(index)
		}
		button.Enable()
	}()
	button.OnClicked(func(*ui.Button) {
		// write service file
		utils.WriteService(&autoservice)
		window2.Destroy()
		os.Exit(0)
	})
	ui.OnShouldQuit(func() bool {
		window2.Destroy()
		return true
	})
}
func main() {
	flag.NewFlagSet("remove", flag.ExitOnError)

	if len(os.Args) == 2 {
		if os.Args[1] == "remove" {
			uninstall()
			return
		}
	}
	usr, err := user.Current()
	if err != nil {
		return
	}
	defaultpath = filepath.Join(usr.HomeDir, "ino-storage")

	if runtime.GOOS == "windows" {
		configpath = "/etc"
		buildpath = "/bin"
		buildpath2 = "/usr/bin"
	} else {
		configpath = "/etc"
		buildpath = "/bin"
		buildpath2 = "/usr/bin"
	}

	ui.Main(setupUI)
}
func uninstall() {
	// remove build from /bin /usr/bin
	// remove program folder from its path
	// remove config file from /etc
	// remove service form /etc/systemd/system
}
func dealwithErr(err error) {
	if err != nil {
		fmt.Println(err)
		//os.Exit(-1)
	}
}
