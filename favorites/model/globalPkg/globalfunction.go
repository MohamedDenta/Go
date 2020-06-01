package globalPkg

import (
	"bytes" //send data into bytes form
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json" //read and send json data through api
	"fmt"
	"io/ioutil" //  read and write on the json file
	"math/rand"

	"../cryptogrpghy"

	"net"
	"net/http" // using API request
	"strconv"
	"strings"
	"time"

	"../errorpk" //  write an error on the json file
	"../logpkg"

	"github.com/dgrijalva/jwt-go"
)

/*----------------function to send data to specific url ----------------- */
func SendRequest(data []byte, url string, method string) string {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return errorpk.AddError("Send Request global function package", "Can't Reach the destinaton ", "critical error")

	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return errorpk.AddError("Send Request global function package", "Can't Reach the destinaton ", "critical error")

	}

	if resp.StatusCode == 500 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return bodyString
	}

	defer resp.Body.Close()
	return ""
}

func SendBroadCast(data []byte, url string, method string) (string, []byte) {
	var t []byte
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return "error", t
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "error", t

	}

	if resp.StatusCode == 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		//	bodyString := string(bodyBytes)
		return "", bodyBytes
	}
	// bodyBytes, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return "error", t
}

/*----------------function to send data to specific url ----------------- */
func SendLedger(data []byte, url string, method string) (string, []byte) {
	var t []byte
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return errorpk.AddError("Send Request global function package", "Can't Reach the destinaton ", "critical error"), t

	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return errorpk.AddError("Send Request global function package", "Can't Reach the destinaton ", "critical error"), t

	}

	if resp.StatusCode == 500 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return bodyString, t
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return "", bodyBytes
}

func SendRequestTransaction(data []byte, url string, method string, objStruct map[string]bool) bool {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("NewRequest", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("client err", err)
		return false
	}

	if resp.StatusCode == 500 {
		// bodyBytes, _ := ioutil.ReadAll(resp.Body)
		// bodyString := string(bodyBytes)
		return false
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &objStruct)

	returnObj := objStruct["isValid"]
	fmt.Println("returnObj", returnObj)

	return objStruct["isValid"]
	// if err != nil {
	// 	return errorpk.AddError("SendRequestAndGetResponse global function package", err.Error())
	// }
	// defer resp.Body.Close()
	// return ""
}

func SendRequestAndGetResponse(data []byte, url string, method string, objStruct interface{}) string {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("NewRequest", err)
		return errorpk.AddError("SendRequestAndGetResponse global function package", "Can't Reach the destinaton ", "critical error")
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("client err", err)
		return errorpk.AddError("SendRequestAndGetResponse global function package", "Can't Reach the destinaton ", "critical error")
	}

	if resp.StatusCode == 500 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return bodyString
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &objStruct)
	objStruct = objStruct
	// fmt.Println("objStruct", objStruct)
	if err != nil {
		return errorpk.AddError("SendRequestAndGetResponse global function package", err.Error(), "critical error")
	}
	defer resp.Body.Close()
	return string(bodyBytes)
}

/*----------function to convert any interface to byte----------*/
func ConvetToByte(data interface{}, funcName string) (value []byte, convert bool) {
	var err error
	value, err = json.Marshal(data)
	if err != nil {
		errorpk.AddError("ConvetToByte "+funcName, "can't convert data to json", "runtime error")
		return value, false
	}
	return value, true
}

/*----------function to convert integar to fixed digits of string----------*/
func ConvertIntToFixedLengthString(key int, length int) (stringform string, err bool) {
	stringform = strconv.Itoa(key)
	stringlen := len(stringform)
	if stringlen > length {
		return "", false
	}
	for i := 0; i < length-stringlen; i++ {
		stringform = "0" + stringform
	}
	return stringform, true
}

/*----------function to convert Convert Fixed Length String to Int----------*/
func ConvertFixedLengthStringtoInt(key string) (stringform int) {
	for index := 0; index < len(key); index++ {
		if key[index:index+1] != "0" {
			number, _ := strconv.Atoi(key[index:len(key)])
			return number
		}

	}
	return 0
}

//SendError status for 503 StatusServiceUnavailable
func SendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	mess := SendMessage{}
	mess.MessageAPI = message
	w.WriteHeader(http.StatusServiceUnavailable)
	sendJSON, _ := json.Marshal(mess)
	w.Write(sendJSON)
}

//SendNotFound status 404 StatusNotFound
func SendNotFound(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")

	mess := SendMessage{}
	mess.MessageAPI = message
	w.WriteHeader(http.StatusNotFound)
	sendJSON, _ := json.Marshal(mess)
	w.Write(sendJSON)
}

//SendResponse status 200 status OK
func SendResponse(w http.ResponseWriter, JSONObj []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(JSONObj)
}
func SendResponseMessage(w http.ResponseWriter, JSONObj string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	mess := SendMessage{}
	mess.MessageAPI = JSONObj
	sendJSON, _ := json.Marshal(mess)
	w.WriteHeader(http.StatusOK)
	w.Write(sendJSON)
}

// Authenticatin and Authorization part

var authAdmin = []byte("Admin key")
var authUser = []byte("User key")

func IsUser(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenStr := r.Header["Jwt-Token"]
		if tokenStr != nil {
			token, err := jwt.Parse(tokenStr[0], func(token *jwt.Token) (interface{}, error) {

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return UserSigningKey, nil
			})
			fmt.Println("error token  --    ", err)
			if err != nil {
				SendError(w, "Not Authorized")
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			SendError(w, "Not Authorized")
		}
	})
}

func IsAdmine(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenStr := r.Header["Jwt-Token"]
		if tokenStr != nil {
			token, err := jwt.Parse(tokenStr[0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return AdminSigningKey, nil
			})
			fmt.Println("error token :    --    ", err)
			if err != nil {
				SendError(w, "Not Authorized")
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			SendError(w, "Not Authorized")
		}
	})
}

// func IsUser(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		if r.Header["Token"] != nil {

// 			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
// 				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 					return nil, fmt.Errorf("There was an error")
// 				}
// 				return authUser, nil
// 			})

// 			if err != nil {
// 				fmt.Fprintf(w, err.Error())
// 				fmt.Fprintf(w, ",   Not Authorized")
// 			}

// 			if token.Valid {
// 				endpoint(w, r)
// 			}
// 		} else {
// 			fmt.Fprintf(w, "Not Authorized")
// 		}
// 	})
// }

// func IsAdmine(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		if r.Header["Token"] != nil {

// 			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
// 				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 					return nil, fmt.Errorf("There was an error")
// 				}
// 				return authAdmin, nil
// 			})

// 			if err != nil {
// 				fmt.Fprintf(w, err.Error())
// 				fmt.Fprintf(w, ",   Not Authorized")
// 			}

// 			if token.Valid {
// 				endpoint(w, r)
// 			}
// 		} else {

// 			fmt.Fprintf(w, "Not Authorized")
// 		}
// 	})
// }

//CreateHash hash math equation for transaction and block
func CreateHash(t time.Time, data string, factor int) string {
	// sum := sha256.Sum256([]byte(data))
	// str := hex.EncodeToString(sum[:])
	h := sha256.New()
	h.Write([]byte(data))
	sum := h.Sum(nil)
	// sum2 := sha256.Sum256([]byte(t.String()))
	// str2 := hex.EncodeToString(sum2[:])
	h2 := sha256.New()
	h2.Write([]byte(t.String()))
	sum2 := h2.Sum(nil)
	var hash [32]byte
	for i := 0; i < len(sum); i++ {
		hash[i] = sum[i] + sum2[i] + byte(factor)
	}
	return hex.EncodeToString(hash[:])
}

//GetHash get hash for ip server
func GetHash(str []byte) string {
	hasher := sha256.New()
	hasher.Write(str)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// TODO: login to service
func serviceLogin(login_url, method string, cred_json []byte) int {
	req, err := http.NewRequest(method, login_url, bytes.NewBuffer(cred_json))
	//log
	now, userIP := SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "serviceLogin", "globalPkg", "_", "_", "_", 0}
	if err != nil {
		WriteLog(logobj, "failed to login", "failed")
		return 1 // error in login credentials
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cookie", "unifises="+CookieObject2[0]+"; csrf_token="+CookieObject2[1])
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		WriteLog(logobj, "timeout , can not reach destination .", "failed")
		return 2 // timout error
	}
	defer resp.Body.Close()
	// read cookies from response login
	for index, cookieObject := range resp.Cookies() {
		CookieObject2[index] = cookieObject.Value
	}
	WriteLog(logobj, "login successfully", "success")
	return 200
}

// CreateVoucher request for create voucher
func CreateVoucher(cred_json, vou_json []byte, login_url, creat_vou_url, method string) (ResponseCreateVoucher, int) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	code := serviceLogin(login_url, method, cred_json)
	if code != 200 {
		return ResponseCreateVoucher{}, 1 // can not login
	}
	//-------------
	req, err := http.NewRequest(method, creat_vou_url, bytes.NewBuffer(vou_json))
	//log
	now, userIP := SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "CreateVoucher", "globalPkg", "_", "_", "_", 0}
	if err != nil {
		WriteLog(logobj, "error in create-voucher request", "failed")
		return ResponseCreateVoucher{}, 3 // error in create-voucher request
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cookie", "unifises="+CookieObject2[0]+"; csrf_token="+CookieObject2[1])
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		WriteLog(logobj, "timeout,can not reach destination", "failed")
		return ResponseCreateVoucher{}, 4 // timout error
	}
	defer resp.Body.Close()
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
			WriteLog(logobj, "empty response"+string(bodyBytes), "failed")
			return ResponseCreateVoucher{}, 5 // empty response
		}
		var ct ResponseCreateVoucher
		e := json.Unmarshal(bh, &ct)
		if e != nil {
			WriteLog(logobj, "empty response"+string(bodyBytes), "failed")
			return ResponseCreateVoucher{}, 5 // empty response
		}
		WriteLog(logobj, "create-voucher response "+string(bodyBytes), "success")
		return getVoucherData(ct)

	}
	WriteLog(logobj, "failed to create service", "failed")
	return ResponseCreateVoucher{}, 0 // not 200 response
}

// TODO: get data for some voucher
func getVoucherData(ct ResponseCreateVoucher) (ResponseCreateVoucher, int) {
	var b interface{}
	var vouch ResponseCreateVoucher
	bd, _ := json.Marshal(ct)
	rq, er := http.NewRequest("POST", GlobalObj.ServiceStatus, bytes.NewBuffer(bd))
	//log
	now, userIP := SetLogObj(rq)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "getVoucherData", "globalPkg", "_", "_", "_", 0}
	if er != nil {
		WriteLog(logobj, "error in request body", "failed")
		return ResponseCreateVoucher{}, 1 // error in request body
	}
	rq.Header.Set("Content-Type", "application/json")
	rq.Header.Set("cookie", "unifises="+CookieObject2[0]+"; csrf_token="+CookieObject2[1])
	client := &http.Client{}
	resp, er := client.Do(rq)
	if er != nil {
		WriteLog(logobj, "timeout , can not reach destination .", "failed")
		return ResponseCreateVoucher{}, 4 // timout error

	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &b)
	m := b.(map[string]interface{})
	d, ok := m["data"].([]interface{})
	if resp.StatusCode == 200 {
		if !ok {
			WriteLog(logobj, "can not read response "+string(bodyBytes), "failed")
			return vouch, 5 // can not read response
		}

		if len(d) > 0 {
			bh, _ := json.Marshal(d[0])
			json.Unmarshal(bh, &vouch)
			WriteLog(logobj, "service info "+string(bh), "success")
			return vouch, 200 // success
		}
	}

	defer resp.Body.Close()
	return vouch, 5
}

//UTCtime return UTC formated time
func UTCtime() time.Time {
	formatedTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
	return formatedTime
}

//UTCtimefield return UTC formated time
func UTCtimefield(timefield time.Time) time.Time {
	formatedTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", timefield.Format("2006-01-02 03:04:05 PM -0000"))
	return formatedTime
}

// get the IP from API request will replace SetLogObj function
func GetIP(req *http.Request)(time.Time, net.IP){
	now := UTCtime()
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	userIP := net.ParseIP(ip)
	return now, userIP
}
// set time ip mac in log object
func SetLogObj(req *http.Request) (time.Time, net.IP) {
	now := UTCtime()
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	userIP := net.ParseIP(ip)
	return now, userIP
}

// set outputdata , process in log object
func WriteLog(logObj logpkg.LogStruct, data, process string) {
	logObj.OutputData = data
	logObj.Process = process
	logpkg.WriteOnlogFile(logObj)
}

// RandomPath random path to url
func RandomPath() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 40
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"
	return str

}

//GenerateJwtToken GetTokenHandler Create jwt token for admin^user
func GenerateJwtToken(username string, isAdmin bool) string {
	/* Create the token */
	token := jwt.New(jwt.SigningMethodHS256)

	/* Create a map to store our claims*/
	claims := token.Claims.(jwt.MapClaims)

	/* Set token claims */
	claims["admin"] = isAdmin
	if len(username) > 0 {
		claims["name"] = username
	} else {
		claims["name"] = "username"
	}
	claims["name"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()
	/* Sign the token with our secret key*/
	var key []byte
	if isAdmin {
		key = AdminSigningKey
	} else {
		key = UserSigningKey
	}
	tokenString, err := token.SignedString(key)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("jwt-token", tokenString)
	return tokenString
}

//Check if err != nill
func Check(err error) {
	if err != nil {
		panic(err)
	}
}

//EncryptECCKeys return encrypted ECC keys with AES
func EncryptECCKeys(Public ecdsa.PublicKey, Private *ecdsa.PrivateKey) (EncryptedPublic string, EncryptedPrivate string) {
	pubInBytes, err1 := json.Marshal(Public)
	Check(err1)
	privInBytes, err2 := json.Marshal(Private)
	Check(err2)

	EncryptedPublic = cryptogrpghy.AESEncrypt(AESEncryptionKey, string(pubInBytes))
	EncryptedPrivate = cryptogrpghy.AESEncrypt(AESEncryptionKey, string(privInBytes))

	return
}
