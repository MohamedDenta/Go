package heartbeat

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"strings"
	"time"

	"../admin"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../serverworkload"
	validator "../validator"
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
func splitHBIPTime(hbDB HeartBeatStruct) (string, string) {
	//split HeartBeatIPTime into IP , Time
	HBIPTime := strings.Split(hbDB.HeartBeatIp_Time, "_")
	HBip, HBtime := HBIPTime[0], HBIPTime[1]
	return HBip, HBtime
}

//convert heartbeat from database into Heartbeat from miner info
func converthbdatabaseTOminerInfo(hbDB HeartBeatStruct) MinersInfo {

	//call split heartbeat IP & time
	HBip, HBtime := splitHBIPTime(hbDB)

	var minfohb MinersInfo
	//miner IP
	minfohb.MinerIP = HBip

	//convert HBtime string into  minfo.TimeStamp Time.time
	//HBtime = time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000")
	// var err error
	minfohb.TimeStamp, _ = time.Parse("2006-01-02 03:04:05 PM -0000", HBtime)
	fmt.Println("----------------  ", minfohb.TimeStamp)
	//miner status
	minfohb.MinerStatus = hbDB.HeartBeatStatus
	minfohb.Workload = hbDB.HeartBeatworkLoad
	return minfohb
}

//convert  Heartbeat from miner info into heartbeat from database
func convertminerInfoTOhbdatabase(minersInfoObj MinersInfo) HeartBeatStruct {

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
	fmt.Println("--------***--------  ", heartBeatObj.HeartBeatIp_Time)
	return heartBeatObj
}

//CompareMinerStatus compare status of miner
func CompareMinerStatus(minersInfoObj MinersInfo, validatorObj validator.ValidatorStruct) {
	validatorObj.ValidatorActive = minersInfoObj.MinerStatus
	validatorObj.ValidatorLastHeartBeat = minersInfoObj.TimeStamp
	validator.UpdateValidator(validatorObj)
	statusHeartBeat := heartBeatStructGetlastPrefix(minersInfoObj.MinerIP)
	//fmt.Println(statusHeartBeat.HeartBeatStatus, minersInfoObj.MinerStatus)
	if statusHeartBeat.HeartBeatIp_Time == "" || statusHeartBeat.HeartBeatStatus != minersInfoObj.MinerStatus {
		heartBeatObj := convertminerInfoTOhbdatabase(minersInfoObj)

		fmt.Println(minersInfoObj.TimeStamp)
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
				fmt.Println("SEND TO ", validatorObj.ValidatorIP)
				nowTime := globalPkg.UTCtime()
				ValidatorLastHeartBeat := validatorObj.ValidatorLastHeartBeat
				diff := nowTime.Sub(ValidatorLastHeartBeat)
				second := int(diff.Seconds())

				if second > 90 {
					CompareMinerStatus(MinersInfo{Message{validatorObj.ValidatorIP, nowTime, "0.0%", false, 0.0, ""}, false}, validatorObj)

				}
			}
		}

	}

}

//GetAllHeartBeat return all hb miner in json obj
func GetAllHeartBeat(w http.ResponseWriter, req *http.Request) {
	//log
	now := globalPkg.UTCtime()
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	userIP := net.ParseIP(ip)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllHeartBeat", "Heartbeat", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		logobj.OutputData = "please enter your correct request"
		logobj.Process = "faild"
		logpkg.WriteOnlogFile(logobj)
		return
	}

	// if Adminobj.AdminUsername == globalPkg.AdminObj.AdminUsername && Adminobj.AdminPassword == globalPkg.AdminObj.AdminPassword {
	if admin.ValidationAdmin(Adminobj) {
		var AllHB []HeartBeatStruct   // slice of hb struct
		var Allminerinfo []MinersInfo //slice of miner info struct

		AllHB = heartBeatStructGetAll() // call get all hb from db
		fmt.Println("******************************************")
		fmt.Println(AllHB)
		fmt.Println("******************************************")

		for _, hb := range AllHB {

			Allminerinfox := converthbdatabaseTOminerInfo(hb)  //call convert into hb db to miner info
			Allminerinfo = append(Allminerinfo, Allminerinfox) //append new hb into miner info

		}

		sendJSON, _ := json.Marshal(Allminerinfo)
		globalPkg.SendResponse(w, sendJSON)
		logobj.OutputData = "get all heartbeat success"
		logobj.Process = "success"
		logpkg.WriteOnlogFile(logobj)
	} else {
		globalPkg.SendError(w, "you are not admin")
		logobj.OutputData = "you are not admin to get heartbeat"
		logobj.Process = "faild"
		logpkg.WriteOnlogFile(logobj)
	}
}

//**********************************************************************************************
// SendHeartBeat send heartbeat
func SendUpdateHeartBeat(UpdateVersion float32, UpdateUrl string) {
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>SendUpdateHeartBeat", UpdateVersion, UpdateUrl)
	//UpdateVersion=UpdateVersion+"\n"
	//updateVersionFloat,_:= strconv.ParseFloat(UpdateVersion,64)
	message := Message{validator.CurrentValidator.ValidatorIP, time.Now().UTC(), serverworkload.GetCPUWorkloadPrecentage(), true, UpdateVersion, UpdateUrl}
	//	compareMinerStatus(minersInfoObj, validator.CurrentValidator)
	broadcastTcp.BoardcastingTCP(message, "", "heartBeat")

}
