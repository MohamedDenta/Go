package service

import (
	"time"

	mathrand "math/rand"

	"../globalPkg"
	//transaction "../transaction"
)

type UserKeys struct {
	PublicKey string
	Password  string
}

type VoucherResponse struct {
	TransactionId string
	Voucher       string
}

type VoucherRequest struct {
	Cmd    string
	N      string
	expire int ///////ask jolin
	up     string
	down   string
	bytes  string
	MBytes bool
	note   string
	quota  string
}

/////-------------
type Vocherstatus struct {
	PublicKey string
	Password  string
	VoucherID string
}

type VocherstatusResponse struct {
	Id          string
	Site_id     string
	Create_time int64
}

//// serviceTemp is adatabase to save all requested
var serviceTemp []ServiceStruct

func CalculateAmountAndCost(serviceobj ServiceStruct) ServiceStruct {
	var M int ///total Amount in megabytes
	if serviceobj.Day == false {
		M = serviceobj.Duration * serviceobj.Bandwidth * 60
	} else {
		M = serviceobj.Duration * serviceobj.Bandwidth * 60 * 24
	}

	C := (float64(M) * 0.001) //+ globalPkg.GlobalObj.TransactionFee
	serviceobj.M = M
	serviceobj.Calculation = C
	return serviceobj
}

////////////////////////////////////////
//------------------------------------------------------------------------------
//   Add service object in serviceTem Array
//  first check if the request found in temp delete it And add the Newest one
//--------------------------------------------------------------------------------
func AddserviceInTmp(serviceobjstruc ServiceStruct) {
	for index, serviceobj := range serviceTemp {
		if serviceobj.PublicKey == serviceobjstruc.PublicKey {
			serviceTemp = append(serviceTemp[:index], serviceTemp[index+1:]...)
			break

		}

	}
	serviceTemp = append(serviceTemp, serviceobjstruc)

}

func AddAndUpdateServiceObj(serviceobj ServiceStruct) {

	ServiceCreateOUpdate(serviceobj)

}
func GetAllservice() []ServiceStruct {

	return serviceTemp

}
func RemoveServicefromTmp(index int) {
	serviceTemp = append(serviceTemp[:index], serviceTemp[index+1:]...)
}
func SetserviceTemp(serviceObject []ServiceStruct) {
	serviceTemp = serviceObject

}
func GetAllPurchusedservice() []ServiceStruct {
	//return serviceTemp
	return ServiceStructGetAll()
}

//------------------------------------------------------------------------------
//   go routine to delete requested servive
//----------------------------------------------------------------
func ClearDeadRequestedService() {
	for {
		time.Sleep(time.Second * time.Duration(mathrand.Int31n(globalPkg.GlobalObj.DeleteAccountLoopTimeInseacond)))

		t := globalPkg.UTCtime()

		for index, serviceobj := range serviceTemp {
			t2 := serviceobj.Time
			Subtime := (t.Sub(t2)).Seconds()
			if Subtime > 3600 { ///globalPkg.GlobalObj.DeleteAccountTimeInseacond {
				serviceTemp = append(serviceTemp[:index], serviceTemp[index+1:]...)

			}
		}
	}
}

//----------------------------------------------------------
//
//----------------------------------------------
func CheckRequestId(ID string, PublicKey string) (ServiceStruct, bool) {
	var EmptyService ServiceStruct
	//servicestructobj = FindServiceById(ID)
	for _, serviceobj := range serviceTemp {
		if serviceobj.PublicKey == PublicKey && serviceobj.ID == ID {
			return serviceobj, true

		}
	}
	return EmptyService, false

}
