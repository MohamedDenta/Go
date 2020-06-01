package heartbeat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"../admin"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../serverworkload"
	validator "../validator"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

//Message struct
type Message struct {
	MinerIP       string
	TimeStamp     time.Time
	Workload      string
	UpdateExist   bool
	UpdateVersion float32
	UpdateUrl     string
}

//MinersInfo struct
type MinersInfo struct {
	Message     //message struct
	MinerStatus bool
}

//split HeartBeatIPTime into IP , Time
func (hbDB *HeartBeatStruct) splitHBIPTime() (string, string) {
	//split HeartBeatIPTime into IP , Time
	HBIPTime := strings.Split(hbDB.HeartBeatIp_Time, "_")
	HBip, HBtime := HBIPTime[0], HBIPTime[1]
	return HBip, HBtime
}

//converthbdatabaseTOminerInfo convert heartbeat from database into Heartbeat from miner info
func (hbDB *HeartBeatStruct) converthbdatabaseTOminerInfo() MinersInfo {
	//call split heartbeat IP & time
	HBip, HBtime := hbDB.splitHBIPTime()
	var minfohb MinersInfo
	//miner IP
	minfohb.MinerIP = HBip

	//convert HBtime string into  minfo.TimeStamp Time.time
	//HBtime = time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000")
	minfohb.TimeStamp, _ = time.Parse("2006-01-02 03:04:05 PM -0000", HBtime)
	// fmt.Println(minfohb.TimeStamp)
	//miner status
	minfohb.MinerStatus = hbDB.HeartBeatStatus
	minfohb.Workload = hbDB.HeartBeatworkLoad
	return minfohb
}

//ConvertminerInfoTOhbdatabase convert  Heartbeat from miner info into heartbeat from database  public to call it in update system
func (minersInfoObj *MinersInfo) ConvertminerInfoTOhbdatabase() HeartBeatStruct {

	var heartBeatObj HeartBeatStruct
	//minerIPTime [] miner ip timestamp
	minerIPTime := []string{minersInfoObj.MinerIP, minersInfoObj.TimeStamp.Format("2006-01-02 03:04:05 PM -0000")}

	//join minerIP & timeStamp
	minerIPTimestring := strings.Join(minerIPTime, "_")
	//put miner ip,time into hb ip,time
	heartBeatObj.HeartBeatIp_Time = minerIPTimestring
	//put miner status into hb status
	heartBeatObj.HeartBeatStatus = minersInfoObj.MinerStatus
	heartBeatObj.HeartBeatworkLoad = minersInfoObj.Workload
	// fmt.Println("--------***--------  ", heartBeatObj.HeartBeatIp_Time)
	return heartBeatObj
}

//CompareMinerStatus compare status of miner if change store it in db
func (minersInfoObj *MinersInfo) CompareMinerStatus(validatorObj validator.ValidatorStruct) {

	validatorObj.ValidatorActive = minersInfoObj.MinerStatus
	validatorObj.ValidatorLastHeartBeat = minersInfoObj.TimeStamp
	(&validatorObj).UpdateValidator()
	statusHeartBeat := heartBeatStructGetlastPrefix(minersInfoObj.MinerIP)
	//fmt.Println(statusHeartBeat.HeartBeatStatus, minersInfoObj.MinerStatus)
	if statusHeartBeat.HeartBeatIp_Time == "" || statusHeartBeat.HeartBeatStatus != minersInfoObj.MinerStatus {
		heartBeatObj := minersInfoObj.ConvertminerInfoTOhbdatabase()

		// fmt.Println(minersInfoObj.TimeStamp)
		heartBeatStructCreate(heartBeatObj)
	}
}

//Network to check all IPs in network
func Network() {
	time.Sleep(60 * time.Second)
	message := Message{validator.CurrentValidator.ValidatorIP, time.Now().UTC(), serverworkload.GetCPUWorkloadPrecentage(), false, 0.0, ""}
	fmt.Println(message)
}

// SendHeartBeat send heartbeat
func SendHeartBeat() {
	for {
		time.Sleep(60 * time.Second)
		message := Message{validator.CurrentValidator.ValidatorIP, time.Now().UTC(), serverworkload.GetCPUWorkloadPrecentage(), false, 0.0, ""}
		//	compareMinerStatus(minersInfoObj, validator.CurrentValidator)
		broadcastTcp.BoardcastingTCP(message, "", "heartBeat")

		for _, validatorObj := range validator.ValidatorsLstObj {
			if validatorObj.ValidatorActive && validatorObj.ValidatorIP != validator.CurrentValidator.ValidatorIP {

				nowTime := globalPkg.UTCtime()
				ValidatorLastHeartBeat := validatorObj.ValidatorLastHeartBeat
				diff := nowTime.Sub(ValidatorLastHeartBeat)
				second := int(diff.Seconds())

				if second > 90 {
					var minfo = MinersInfo{Message{validatorObj.ValidatorIP, nowTime, "0.0%", false, 0.0, ""}, false}
					minfo.CompareMinerStatus(validatorObj)

				}
			}
		}

	}

}

//GetAllHeartBeat return all hb miner
func GetAllHeartBeat(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllHeartBeat", "Heartbeat", "_", "_", "_", 0}
	if !admin.AdminAPIDecoderAndValidation(w, req.Body, logobj) {
		return
	}

	var AllHB []HeartBeatStruct   // slice of hb struct
	var Allminerinfo []MinersInfo //slice of miner info struct

	AllHB = heartBeatStructGetAll() // call get all hb from db

	// fmt.Println(AllHB)

	for _, hb := range AllHB {
		Allminerinfox := hb.converthbdatabaseTOminerInfo() //call convert into hb db to miner info
		Allminerinfo = append(Allminerinfo, Allminerinfox) //append new hb into miner info
	}

	sendJSON, _ := json.Marshal(Allminerinfo)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get all heartbeat success", "success")
}

//SendUpdateHeartBeat send heartbeat to update system
func SendUpdateHeartBeat(UpdateVersion float32, UpdateUrl string) {
	// fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>SendUpdateHeartBeat", UpdateVersion, UpdateUrl)
	//UpdateVersion=UpdateVersion+"\n"
	//updateVersionFloat,_:= strconv.ParseFloat(UpdateVersion,64)
	message := Message{validator.CurrentValidator.ValidatorIP, time.Now().UTC(), serverworkload.GetCPUWorkloadPrecentage(), true, UpdateVersion, UpdateUrl}
	//	compareMinerStatus(minersInfoObj, validator.CurrentValidator)
	broadcastTcp.BoardcastingTCP(message, "", "heartBeat")

}

//CreateHeartbeatAfterSystemdown create hb after system down
func CreateHeartbeatAfterSystemdown(hbDB HeartBeatStruct) {

	heartBeatStructCreate(hbDB)
}

// BoardcastHandleHeartbeat handle heartbeat case
func BoardcastHandleHeartbeat(tCPDataObj broadcastTcp.TCPData) {
	var message Message
	var heartbeatObjec MinersInfo
	// fmt.Println("the local server is := ", conn.LocalAddr())
	// fmt.Println("the other servers server are := ", conn.RemoteAddr())
	message.TimeStamp = globalPkg.UTCtime()
	// mapstructure.Decode(tCPDataObj.Obj, &message)
	json.Unmarshal(tCPDataObj.Obj, &message)
	fmt.Println(message)
	if message.UpdateExist == false {

		heartbeatObjec.MinerStatus = true
		heartbeatObjec.Message = message
		fmt.Println(message)
		for _, validatorObj := range validator.ValidatorsLstObj {
			if validatorObj.ValidatorIP == heartbeatObjec.Message.MinerIP {
				// heartbeat.CompareMinerStatus(heartbeatObjec, validatorObj)
				heartbeatObjec.CompareMinerStatus(validatorObj)
				break
			}
		}
	} else {
		viper.SetConfigName("./config")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig() // Find and read the config file

		if err != nil { // Handle errors reading the config file
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
		var currentVersion float32
		value, _ := strconv.ParseFloat(viper.GetString("updatestruct.currentversion"), 32)
		currentVersion = float32(value)
		viper.Set("updatestruct.updateversion", message.UpdateVersion)
		viper.Set("updatestruct.updateurl", message.UpdateUrl)
		viper.WriteConfig()

		if message.UpdateVersion > currentVersion {
			fmt.Println("update recieved")
			viper.Set("updatestruct.currentversion", message.UpdateVersion)
			viper.WriteConfig()

			fmt.Println("\n downloading update version : ", message.UpdateVersion)
			err = DownloadFile(message.UpdateUrl)

			delaySecond(60)

			cmd := exec.Command("sudo", "systemctl", "start", "auto.service")
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			out, err := cmd.Output()
			if err != nil {
				fmt.Println("Err", err)
			} else {
				fmt.Println("OUT:", string(out))
			}

		}
	}
}

//WriteCounter write counter
type WriteCounter struct {
	Total uint64
}

//delaySecond delay second
func delaySecond(n time.Duration) {
	time.Sleep(n * time.Second)
}

//Write write
func (WCount *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	WCount.Total += uint64(n)
	WCount.PrintProgress()
	return n, nil
}

//PrintProgress print progress
func (WCount WriteCounter) PrintProgress() {

	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	//print bytes in a meaningful way
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(WCount.Total))
}

//DownloadFile download file for system update
func DownloadFile(url string) error {

	//download file .tmp file
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
	return nil
}
