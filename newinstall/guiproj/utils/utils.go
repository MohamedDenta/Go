package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"newinstall/guiproj/deviceinfo"
	"newinstall/guiproj/hide"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
)

// readLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func ReadLines(filename string) ([]string, error) {
	return ReadLinesOffsetN(filename, 0, -1)
}

// readLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func ReadLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

func PathExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// stringsHas checks the target string slice contains src or not
func StringsHas(target []string, src string) bool {
	for _, t := range target {
		if strings.TrimSpace(t) == src {
			return true
		}
	}
	return false
}

// formatBytes format the
func FormatBytes(bytes uint64, decimals uint8) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	const k float64 = 1000
	var sizes = [...]string{"Bytes", "KB", "MB", "GB", "TB", "PB"}

	var i float64 = math.Floor(math.Log(float64(bytes)) / math.Log(float64(k)))

	return fmt.Sprintf(fmt.Sprintf("%%.%vf %%s", i), float64(bytes)/math.Pow(k, i), sizes[int(i)])
}

// Respond with Json data
func Respond(w http.ResponseWriter, status int, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteConfigFile write to config file
func WriteConfigFile(deviceinfobj *(deviceinfo.Info), configpath *string) (string, bool) {
	var fname string = "ino-config.conf"
	// if runtime.GOOS == "linux" {
	// 	fname = "." + fname
	// }
	// write resaurces to config file
	filename := filepath.Join(*configpath, fname)

	newFile, err := os.Create(fname)
	if err != nil {
		fmt.Println(err)
		return "error in create file", false
	}

	//newFile.Chmod(0700) // read and write execute for only owner
	b, _ := json.Marshal(deviceinfobj)
	// err = hide.HideFile(filename)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return "error in hide file", false
	// }
	err = ioutil.WriteFile(fname, b, 0600)
	newFile.Close()
	cmd := exec.Command("/bin/sh", "-c", "sudo mv "+fname+" "+filename)
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
	cmd = exec.Command("/bin/sh", "-c", " sudo cp inostorage /bin /usr/bin")
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

// ReserveStorage create a file with a specified size
func ReserveStorage(storage int64) error {
	filename := `storagenode.ino`
	os.Remove(filename)
	os.Remove("." + filename)
	filename = filepath.Join(Defaultpath, filename)
	newFile, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return errors.New("failed to create storage")
	}
	defer newFile.Close()
	newFile.Chmod(0700) // read and write execute for only owner
	newFile.Truncate(storage)
	err = hide.HideFile(filename)
	if err != nil {
		fmt.Println(err)
		return errors.New("failed to hide storage")
	}
	fmt.Println("hidden:", filename)
	return nil
}
