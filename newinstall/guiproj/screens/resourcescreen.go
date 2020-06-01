package screens

import (
	"fmt"
	"newinstall/guiproj/deviceinfo"
	"newinstall/guiproj/net"
	"newinstall/guiproj/utils"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

func dealwithErr(err error) {
	if err != nil {
		fmt.Println(err)
		//os.Exit(-1)
	}
}

//BuildContent build first screen content
func ResourcesScreen() *fyne.Container {
	var deviceinfobj deviceinfo.Info
	ret := fyne.NewContainerWithLayout(layout.NewGridLayout(2))
	cpucores := deviceinfo.Getcpucores()
	vmStat, err := mem.VirtualMemory()
	dealwithErr(err)
	hostStat, _ := host.Info()
	diskStat, err := disk.Usage("/")
	dealwithErr(err)
	cpuStat, err := cpu.Info()
	dealwithErr(err)
	_, err = cpu.Percent(0, true)
	dealwithErr(err)
	var coresspeedlbl []*widget.Label
	lbpath := widget.NewLabel("cpu cores : " + strconv.Itoa(cpucores))
	for i, cpu := range cpuStat {
		deviceinfobj.CpuInfo.CpuSpeed = append(deviceinfobj.CpuInfo.CpuSpeed, strconv.Itoa(i)+strconv.FormatFloat(cpu.Mhz, 'f', 2, 64))
		coresspeedlbl = append(coresspeedlbl, widget.NewLabel("Speed cpu : "+strconv.Itoa(i)+" "+strconv.Itoa(i)+strconv.FormatFloat(cpu.Mhz, 'f', 2, 64)+" MHz"))
	}
	ostypelbl := widget.NewLabel("OS Type : " + hostStat.OS)
	ret.AddObject(ostypelbl)

	ospltfrmlbl := widget.NewLabel("OS Platform : " + hostStat.Platform)
	ret.AddObject(ospltfrmlbl)

	totalmemlbl := widget.NewLabel("total mem : " + strconv.FormatUint(vmStat.Total, 10))
	deviceinfobj.MemoryInfo.TotalMemory = strconv.FormatUint(vmStat.Total, 10)
	ret.AddObject(totalmemlbl)
	fmt.Println(vmStat)
	freememlbl := widget.NewLabel("free mem : " + strconv.FormatUint(vmStat.Available, 10))
	deviceinfobj.MemoryInfo.FreeMemory = strconv.FormatUint(vmStat.Available, 10)
	ret.AddObject(freememlbl)

	usedpercentmemlbl := widget.NewLabel("used percent mem : " + strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64) + "%")
	deviceinfobj.MemoryInfo.PercentageMemory = strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64)
	ret.AddObject(usedpercentmemlbl)

	totaldiskspacelbl := widget.NewLabel("Total disk space: " + strconv.FormatUint(diskStat.Total, 10) + " bytes")
	deviceinfobj.DiskInfo.TotalDiskSpace = strconv.FormatUint(diskStat.Total, 10)
	ret.AddObject(totaldiskspacelbl)

	useddiskspacelbl := widget.NewLabel("Used disk space: " + strconv.FormatUint(diskStat.Used, 10) + " bytes")
	deviceinfobj.DiskInfo.UsedDiskSpace = strconv.FormatUint(diskStat.Used, 10)
	ret.AddObject(useddiskspacelbl)

	freediskspacelbl := widget.NewLabel("Free disk space: " + strconv.FormatUint(diskStat.Free, 10) + " bytes")
	deviceinfobj.DiskInfo.FreeDiskSpace = strconv.FormatUint(diskStat.Free, 10)
	ret.AddObject(freediskspacelbl)

	percentageusedisklbl := widget.NewLabel("Percentage disk space usage: " + strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64) + "%")
	deviceinfobj.DiskInfo.PercentageDiskSpace = strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64)
	ret.AddObject(percentageusedisklbl)

	networkStats, _ := net.IOCounters()
	var netlbl []*widget.Label
	for _, counter := range networkStats {
		// ifcounter.Name=="wlp2s0")
		// TODO: check interface name
		netlbl = append(netlbl, widget.NewLabel("name  "+counter.Name))
		ret.AddObject(widget.NewLabel("name  " + counter.Name))

		netlbl = append(netlbl, widget.NewLabel("bytesSent "+utils.FormatBytes(counter.BytesSent, 2)))
		ret.AddObject(widget.NewLabel("bytesSent " + utils.FormatBytes(counter.BytesSent, 2)))

		netlbl = append(netlbl, widget.NewLabel("bytesRecv "+utils.FormatBytes(counter.BytesRecv, 2)))
		ret.AddObject(widget.NewLabel("bytesRecv " + utils.FormatBytes(counter.BytesRecv, 2)))

		ret.AddObject(widget.NewLabel("packetsSent " + utils.FormatBytes(counter.PacketsSent, 2)))
		netlbl = append(netlbl, widget.NewLabel("packetsSent "+utils.FormatBytes(counter.PacketsSent, 2)))

		ret.AddObject(widget.NewLabel("packetsRecv " + utils.FormatBytes(counter.PacketsRecv, 2)))
		netlbl = append(netlbl, widget.NewLabel("packetsRecv "+utils.FormatBytes(counter.PacketsRecv, 2)))

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
	iplbl := widget.NewLabel("IP : " + ip)
	ret.AddObject(iplbl)

	maclbl := widget.NewLabel("MAC : " + mac)
	ret.AddObject(maclbl)

	cont := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	//	ret.Layout.Layout()

	ret.AddObject(lbpath)

	cont.AddObject(ret)
	cont.AddObject(layout.NewSpacer())
	row := widget.NewGroup("",
		fyne.NewContainerWithLayout(layout.NewGridLayout(2),

			widget.NewButton("Back", func() {
				Wind.SetContent(WelcomeScreen())

			}),
			widget.NewButton("Next", func() {
				deviceinfobj.StoragePath = utils.Defaultpath
				str, isfine := utils.WriteConfigFile(&deviceinfobj, &utils.Configpath)
				if !isfine {
					fmt.Println(str)
				}
				Wind.SetContent(ServiceScreen())
			}),
		),
	)

	cont.AddObject(row)
	return cont
}
