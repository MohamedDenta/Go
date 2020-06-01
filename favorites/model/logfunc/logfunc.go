package logfunc

import (
	"fmt"

	"../logpkg"

	//	"log"
	"net"
	//"os"
	"strings"
	//"time"

	"../globalPkg"
)

//-------------------------------------------
//   func to write on log file
//----------------------111111111111111----------------------------

func WriteAndUpdateLog(logStructObject logpkg.LogStruct) {
	now := globalPkg.UTCtime()
	logStructObject.Currenttime = now
	if !logpkg.RecordLog(logStructObject) {
		fmt.Println("Cant not write log to db")
	}

}
func getLastIndex() string {

	var logobj logpkg.LogStruct
	logobj = logpkg.GetLastlogObj()
	//if Account.AccountPublicKey == "" {
	//	return "-1"
	//}
	if logobj.IPA == nil {
		return "-1"
	}

	return logobj.Index

}

////////////////////////////////---------------

func NewLogIndex() string {
	LastIndex := getLastIndex()
	//fmt.Println(" ----------- last  index---------       " , LastIndex)
	index := 0
	if LastIndex != "-1" {
		// TODO : split LastIndex
		res := strings.Split(LastIndex, "_")

		index = globalPkg.ConvertFixedLengthStringtoInt(res[1]) + 1

	}
	timpIndex, _ := globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.StringFixedLength)

	return timpIndex
}

func CheckIPBlocked(IP net.IP) bool {
	logobj := logpkg.GetLogByIP(IP)
	if logobj.Count <= 10 {
		return true
	}

	return false

}

func GetAllLogs() []logpkg.LogStruct {
	return logpkg.GetAllLogsFromDB()
}

//-----------------------------------
//  Replace the logobj
//--------------------------------------------------

func ReplaceLog(logobj logpkg.LogStruct, funcName string, moduleName string) logpkg.LogStruct {

	if logobj.Function != funcName {
		logobj.Function = funcName
		logobj.Module = moduleName

	}
	return logobj

}
