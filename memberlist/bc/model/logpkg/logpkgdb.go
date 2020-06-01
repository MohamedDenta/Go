package logpkg

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"net"
	"os"

	//	"../errorpk"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	Now time.Time

	outfile, _ = os.Create("my.log") // update path for your needs
	l          = log.New(outfile, "", 0)
)

type LogStruct struct {
	Index       string
	Currenttime time.Time
	IPA         net.IP
	MacAddress  string
	Function    string
	Module      string
	InputData   string
	OutputData  string
	Process     string /////faild or success
	Count       int
}

var (
	DB        *leveldb.DB
	Open      = false
	logtxt, _ = os.Create("log-db.txt") // update path for your needs
	lg        = log.New(logtxt, "", 0)
)

func Opendatabase() bool {
	fmt.Println("open", Open)
	if !Open {
		Open = true
		dbpath := "Database/LoggerDB"
		var err error
		DB, err = leveldb.OpenFile(dbpath, nil)

		if err != nil {

			//errorpk.AddError("opendatabase LoggerStruct package", "can't open the database", "critical error")
			return false
		}
	}

	return true

}

/*----------function to convert any interface to byte----------*/
func ConvetToByte(data interface{}, funcName string) (value []byte, convert bool) {
	var err error
	value, err = json.Marshal(data)
	if err != nil {
		//errorpk.AddError("ConvetToByte "+funcName, "can't convert data to json", "runtime error")
		return value, false
	}
	return value, true
}

func RecordLog(data LogStruct) bool {
	//fmt.Println("+^+ ", data)

	Opendatabase()
	var err error
	d, convert := ConvetToByte(data, "accountCreate account package")
	if !convert {
		return false
	}

	err = DB.Put([]byte(data.Index), d, nil)
	if err != nil {
		//	errorpk.AddError("ServiceStructCreate  ServiceStruct package", "can't create ServiceStruct", "runtime error")

		return false
	}
	closedatabase()
	return true
}
func GetAllLogsFromDB() (values []LogStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata LogStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}
func GetLogByIP(IPAddress net.IP) (values LogStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	fmt.Println("............", values)
	for iter.Next() {

		value := iter.Value()
		fmt.Println("............", values)
		var newdata LogStruct
		json.Unmarshal(value, &newdata)
		if newdata.IPA.Equal(IPAddress) {
			fmt.Println("ipppppppppA", IPAddress, "klklklklk", newdata.IPA)
			values = newdata
			break
		}

	}
	closedatabase()
	fmt.Println("ipppppppppA", values)

	return values

}

//-----------------------------------------------------
//      check if userfound in log data base
//---------------------------------------------------

func CheckIfLogFound(ip net.IP) (bool, LogStruct) {
	logobj := GetLogByIP(ip)

	if logobj.IPA == nil {
		return false, logobj
	}
	return true, logobj
}

//---------------------get las log file
func GetLastlogObj() LogStruct {
	Opendatabase()
	var result LogStruct
	iter := DB.NewIterator(nil, nil)
	for iter.Last() {
		value := iter.Value()
		//fmt.Println("******     value   ", value )
		json.Unmarshal(value, &result)
		break
	}

	//fmt.Println("   ------------ result  --       "   , result)
	closedatabase()
	return result
}
func closedatabase() bool {
	// var err error
	// err = DB.Close()
	// if err != nil {
	// 	//errorpk.AddError("closedatabase LoggerStruct package", "can't close the database")
	// 	return false
	// }
	return true
}
func WriteOnlogFile(logStructObject LogStruct) {

	//path := "my.log"

	//outfile, _ = os.Create(path) // update path for your needs
	//l = log.New(outfile, "", 0)
	l.Print(logStructObject.Currenttime, "  ", logStructObject.Function, "  ", logStructObject.InputData, "  ", logStructObject.OutputData, "  ", logStructObject.Process, "\n")

	// l.Print(logStructObject.IPA)
	// l.Print(logStructObject.MacAddress)
	// l.Print(logStructObject.Function)
	// l.Print(logStructObject.InputData)
	//l.Print(logStructObject.OutputData)

	//l.Println(logStructObject.Process)
}
