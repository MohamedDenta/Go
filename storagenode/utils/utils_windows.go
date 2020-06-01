// +build windows

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"storagenode/deviceinfo"
	"storagenode/hide"

	"golang.org/x/sys/windows"
)

var (
	Modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	ModPsapi    = windows.NewLazySystemDLL("psapi.dll")
)

// WriteConfigFile write to config file
func WriteConfigFile(deviceinfobj *deviceinfo.Info, configpath *string) (string, bool) {
	var fname string = "ino-config.conf"
	// write resaurces to config file
	filename := filepath.Join(*configpath, fname)

	newFile, err := os.Create(fname)
	if err != nil {
		fmt.Println(err)
		return "error in create file", false
	}
	defer newFile.Close()

	//newFile.Chmod(0700) // read and write execute for only owner
	b, _ := json.Marshal(deviceinfobj)
	err = hide.HideFile(filename)
	if err != nil {
		fmt.Println(err)
		return "error in hide file", false
	}
	err = ioutil.WriteFile(fname, b, 0700)
	cmd := exec.Command("move  " + fname + " " + filename)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "error in move config file", false
	}
	return "", true
}

// WriteService write auto-service
func WriteService(autoservice *string) (string, bool) {
	serv, err := os.Create("ino.service")
	if err != nil {
		fmt.Println(err)
		return "error in create service", false
	}
	defer serv.Close()
	n, err := serv.WriteString(*autoservice)
	if err != nil {
		fmt.Println(err)
		return "error in write service", false
	}
	fmt.Println("### ", n)
	cmd := exec.Command("/bin/sh", "-c", " sudo mv ino.service /etc/systemd/system/")
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "error in write service", false
	}
	cmd = exec.Command("/bin/sh", "-c", " sudo cp install-demo /bin /usr/bin")
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "error in write service", false
	}
	cmd = exec.Command("/bin/sh", "-c", " sudo systemctl start ino.service")
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		return "error in start service", false
	}
	return "", true
}
