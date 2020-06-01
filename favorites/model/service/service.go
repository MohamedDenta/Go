package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	mathrand "math/rand"
	"net/http"
	"time"

	"../globalPkg"
	"../logpkg"
)

// serviceTemp is a database to save all requested
var serviceTemp []ServiceStruct

// CalculateAmountAndCost calculate amount and cost
func (serviceobj *ServiceStruct) CalculateAmountAndCost() {
	var M int ///total Amount in megabytes
	if serviceobj.Day == false {
		M = serviceobj.Duration * serviceobj.Bandwidth * 60
	} else {
		M = serviceobj.Duration * serviceobj.Bandwidth * 60 * 24
	}

	C := (float64(M) * 0.001) //+ globalPkg.GlobalObj.TransactionFee
	serviceobj.M = M
	serviceobj.Calculation = C
}

// AddserviceInTmp object in serviceTem Array
// first check if the request found in temp delete it And add the Newest one
func (serviceobjstruc *ServiceStruct) AddserviceInTmp() {
	for index, serviceobj := range serviceTemp {
		if serviceobj.PublicKey == serviceobjstruc.PublicKey {
			serviceTemp = append(serviceTemp[:index], serviceTemp[index+1:]...)
			break
		}
	}
	serviceTemp = append(serviceTemp, *serviceobjstruc)
}

//AddAndUpdateServiceObj add , update service object
func (serviceobj *ServiceStruct) AddAndUpdateServiceObj() {
	serviceobj.ServiceCreateOUpdate()
}

//GetAllservice get all services
func GetAllservice() []ServiceStruct {
	return serviceTemp
}

//RemoveServicefromTmp remove service from temp array
func RemoveServicefromTmp(index int) {
	serviceTemp = append(serviceTemp[:index], serviceTemp[index+1:]...)
}

//SetserviceTemp set temp array
func SetserviceTemp(serviceObject []ServiceStruct) {
	serviceTemp = serviceObject
}

//GetAllPurchusedservice get all purchased service
func GetAllPurchusedservice() []ServiceStruct {
	//return serviceTemp
	return ServiceStructGetAll()
}

//ClearDeadRequestedService go routine to delete service requeste
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

// CheckRequestID check if service request exists in temp array or not
func CheckRequestID(ID string, PublicKey string) (ServiceStruct, bool) {
	var EmptyService ServiceStruct
	for _, serviceobj := range serviceTemp {
		if serviceobj.PublicKey == PublicKey && serviceobj.ID == ID {
			return serviceobj, true
		}
	}
	return EmptyService, false
}

// ServiceLogin login to service
func ServiceLogin(login_url, method string, cred_json []byte) int {
	req, err := http.NewRequest(method, login_url, bytes.NewBuffer(cred_json))
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "serviceLogin", "globalPkg", "", "", "", 0}
	if err != nil {
		globalPkg.WriteLog(logobj, "failed to login", "failed")
		return 1 // error in login credentials
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cookie", "unifises="+globalPkg.CookieObject2[0]+"; csrf_token="+globalPkg.CookieObject2[1])
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		globalPkg.WriteLog(logobj, "timeout , can not reach destination .", "failed")
		return 2 // timout error
	}

	// read cookies from response login
	for index, cookieObject := range resp.Cookies() {
		globalPkg.CookieObject2[index] = cookieObject.Value
	}
	globalPkg.WriteLog(logobj, "login successfully to service", "success")
	defer resp.Body.Close()
	return 200
}

// CreateVoucher request for create voucher
func CreateVoucher(cred_json, vou_json []byte, login_url, creat_vou_url, method string) (ResponseCreateVoucher, int) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
CREATE_VOUCHER:
	req, err := http.NewRequest(method, creat_vou_url, bytes.NewBuffer(vou_json))
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "CreateVoucher", "service", "", "", "", 0}
	if err != nil {
		globalPkg.WriteLog(logobj, "error in create-voucher request", "failed")
		return ResponseCreateVoucher{}, 3 // error in create-voucher request
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cookie", "unifises="+globalPkg.CookieObject2[0]+"; csrf_token="+globalPkg.CookieObject2[1])
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		globalPkg.WriteLog(logobj, "timeout,can not reach destination", "failed")
		return ResponseCreateVoucher{}, 4 // timout error
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		code := ServiceLogin(login_url, method, cred_json)
		if code != 200 {
			return ResponseCreateVoucher{}, 1 // can not login
		}
		goto CREATE_VOUCHER
	}
	if resp.StatusCode == 200 {
		var b interface{}
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(bodyBytes, &b)
		m := b.(map[string]interface{})
		d, _ := m["data"].([]interface{})
		var bh []byte
		if len(d) > 0 {
			bh, _ = json.Marshal(d[0])
		} else {
			globalPkg.WriteLog(logobj, "empty response"+string(bodyBytes), "failed")
			return ResponseCreateVoucher{}, 5 // empty response
		}
		var ct ResponseCreateVoucher
		e := json.Unmarshal(bh, &ct)
		if e != nil {
			globalPkg.WriteLog(logobj, "empty response"+string(bodyBytes), "failed")
			return ResponseCreateVoucher{}, 5 // empty response
		}
		globalPkg.WriteLog(logobj, "create-voucher response "+string(bodyBytes), "success")
		return GetVoucherData(ct)

	}
	globalPkg.WriteLog(logobj, "failed to create service", "failed")
	return ResponseCreateVoucher{}, 0 // not 200 response
}

// GetVoucherData get data for some voucher
func GetVoucherData(ct ResponseCreateVoucher) (ResponseCreateVoucher, int) {
	var b interface{}
	var vouch ResponseCreateVoucher
	bd, _ := json.Marshal(ct)
	rq, er := http.NewRequest("POST", globalPkg.GlobalObj.ServiceStatus, bytes.NewBuffer(bd))
	//log
	now, userIP := globalPkg.SetLogObj(rq)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "GetVoucherData", "globalPkg", "", "", "", 0}
	if er != nil {
		globalPkg.WriteLog(logobj, "error in request body", "failed")
		return ResponseCreateVoucher{}, 1 // error in request body
	}
	rq.Header.Set("Content-Type", "application/json")
	rq.Header.Set("cookie", "unifises="+globalPkg.CookieObject2[0]+"; csrf_token="+globalPkg.CookieObject2[1])
	client := &http.Client{}
	resp, er := client.Do(rq)
	if er != nil {
		globalPkg.WriteLog(logobj, "timeout , can not reach destination .", "failed")
		return ResponseCreateVoucher{}, 4 // timout error

	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ResponseCreateVoucher{}, http.StatusUnauthorized
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &b)
	m := b.(map[string]interface{})
	d, ok := m["data"].([]interface{})
	if resp.StatusCode == 200 {
		if !ok {
			globalPkg.WriteLog(logobj, "can not read response "+string(bodyBytes), "failed")
			return vouch, 5 // can not read response
		}

		if len(d) > 0 {
			bh, _ := json.Marshal(d[0])
			json.Unmarshal(bh, &vouch)
			globalPkg.WriteLog(logobj, "service info "+string(bh), "success")
			return vouch, 200 // success
		}
	}

	defer resp.Body.Close()
	return vouch, 5
}
