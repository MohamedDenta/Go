package systemupdate

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/BurntSushi/toml"
	//"os/exec"
	"strings"
	"time"
	//"strconv"

	"../heartbeat"
	"github.com/dustin/go-humanize"
)

type UpdateData struct {
	Updatestruct Updatestruct
}
type Updatestruct struct {
	Currentversion float32
	Updateversion  float32
	Updateurl      string
}

func delaySecond(n time.Duration) {
	time.Sleep(n * time.Second)
}

type WriteCounter struct {
	Total uint64
}

func (WCount *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	WCount.Total += uint64(n)
	WCount.PrintProgress()
	return n, nil
}

func (WCount WriteCounter) PrintProgress() {

	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	//print bytes in a meaningful way
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(WCount.Total))
}

func DownloadFile(filepath string, url string) error {

	//download file .tmp file and remove .tmp extension when finnished
	name := "build"
	out, err := os.Create(name + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// progress reporter alongside writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// print new line when downloading finnished
	fmt.Print("\n")

	//err = os.Rename(name+".tmp", filepath)
	//if err != nil {
	//	return err
	//}

	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Config map[string]string

func ReadFile(filename string) (Config, error) {
	// init with some bogus data
	config := Config{
		"updateversion2": "1.2",
		"updateurl2":     "https://upload.wikimedia.org/wikipedia/commons/d/d6/Wp-w4-big.jpg",
	}
	if len(filename) == 0 {
		return config, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')

		// check if the line has = sign
		// and process the line. Ignore the rest.
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				// assign the config map
				config[key] = value
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func Update() {
	//fmt.Println("checking update every 3 seconds")
	for {
		//delaySecond(1800)
		delaySecond(60)
		fmt.Println("cheking update !!")

		//time.Sleep(9 * time.Second)
		fmt.Println(">>>>")
		var UpdateDataObj UpdateData
		toml.DecodeFile("./config.toml", &UpdateDataObj)
		fmt.Println(">>>>", UpdateDataObj.Updatestruct.Currentversion)
		fmt.Println(">>>>", UpdateDataObj.Updatestruct.Updateversion)

		// obj1 := Updatestruct{}
		// toml.DecodeFile("./././update.toml", obj1)
		// fmt.Println(">>>>>>>>.", obj1)

		if UpdateDataObj.Updatestruct.Updateversion > UpdateDataObj.Updatestruct.Currentversion {
				heartbeat.SendUpdateHeartBeat(UpdateDataObj.Updatestruct.Updateversion, UpdateDataObj.Updatestruct.Updateurl)
				// changes in config file where current version = update version
				fmt.Println("new update found")
			
			 	
			// 	fmt.Println("updateurl :", UpdateDataObj.Updatestruct.Updateurl)
		// 	UpdateDataObj.Updatestruct.Currentversion =  UpdateDataObj.Updatestruct.Updateversion
		// 	fmt.Println("killing old build ")
		// 	cmd := exec.Command("pkill","htop")
		// 	fmt.Println("Running command and waiting for it to finish...")
		// 	err := cmd.Run()
		// 	fmt.Println("we can't close build because", err)

		// err = DownloadFile("/home/emad/",UpdateDataObj.Updatestruct.Updateurl )

		// if err != nil {
		// 	// ...
		// } else {
		// 	// ...
		// }
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println("Download Complete")

	} 
		// hard code it to config.txt
		// config, err := ReadConfig("./././update.toml")

		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		//fmt.Println("Config data dump :", config)

		// assign values from config file to variables

		// updateversion := config["1.2"]
		// currentversion := config["1.0"]
		// updateurl := config["updateurl"]
		// fmt.Println(">>>>>>>>.", updateversion)

		// config2, err := ReadFile(`./././file.txt`)
		// //
		// if err != nil {
		// 	fmt.Println(err)
		// }

		//fmt.Println("Config data dump :", config)

		// assign values from config file to variables

		// updateversion2 := config2["1.2"]
		// updateurl2 := config2["updateurl"]

		// fmt.Println("currentversion :", currentversion)
		// fmt.Println("updateversion2 :", updateversion2)
		// fmt.Println("updateurl :", updateurl2)
		//time := 0
		//for time == 0 {
	

		
		// 	// d1 := []byte(updateurl)
		// 	// err = ioutil.WriteFile("/home/emadhany/Desktop/emad.txt", d1, 0644)
		// 	// check(err)

		// 	// } else if updateversion2 > currentversion {

		// 	// 	fmt.Println("new update found")
		// 	// 	fmt.Println("\n downloading update version ", updateversion)
		// 	// 	fmt.Println("updateurl :", updateurl)
		// 	// 	err = DownloadFile("/home/emadhany/Desktop", updateurl)

		// 	// 	pass := "emadhany2007"
		// 	// 	_, err = exec.Command("sh", "-c", "echo '"+pass+"' | sudo -S pkill -SIGINT htop").Output()

		// 	// 	err = DownloadFile("/home/emadhany/Desktop", updateurl2)

		// 	// 	// if err != nil {
		// 	// 	// 	// ...
		// 	// 	// } else {
		// 	// 	// 	// ...
		// 	// 	// }
		// 	// 	if err != nil {
		// 	// 		panic(err)
		// 	// 	}
		// 	// 	fmt.Println("Download Complete")
		// 	// 	//break
		// } else {
		// 	fmt.Println("no update found !!")

		// }
		//}

	}
}
