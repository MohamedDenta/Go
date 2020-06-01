package broadcastTcp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"../cryptogrpghy"
	"../globalPkg"
	"../validator"
	"github.com/mitchellh/mapstructure"
)

func sendToLocalNode(obj *interface{}, validatorObj *validator.ValidatorStruct, PackageName, Method string) (resp TCPData, res TxBroadcastResponse, resFile FileBroadcastResponse) {
	if PackageName == "transaction" && Method == "addTransaction" {

		_, res, _ = SendObject(*obj, validatorObj.ValidatorPublicKey, Method, PackageName, validator.CurrentValidator.ValidatorSoketIP)

	} else if PackageName == "file" && Method == "getchunkdata" {

		_, _, resFile = SendObject(*obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

	} else if PackageName == "file" && Method == "addchunk" {

		_, _, resFile = SendObject(*obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

	} else {

		SendObject(*obj, validatorObj.ValidatorPublicKey, Method, PackageName, validator.CurrentValidator.ValidatorSoketIP)

		m := new(MyMessage)
		m.Value, _ = json.Marshal(*obj)

		host, port, err := net.SplitHostPort(validator.CurrentValidator.ValidatorSoketIP)
		host2, port2, _ := net.SplitHostPort(validatorObj.ValidatorSoketIP)

		ip := net.ParseIP(host)
		m.FromAddr = ip
		x, _ := strconv.Atoi(port)
		m.FromPort = uint16(x)

		ip = net.ParseIP(host2)
		m.Key = validatorObj.ValidatorPublicKey
		x, _ = strconv.Atoi(port2)
		i := uint(x)
		err = sendToClient(m, host2, i)
		fmt.Println("$) ", err)

	}
	return resp, res, resFile
}
func sendToNode(obj *interface{}, validatorObj *validator.ValidatorStruct, PackageName, Method string) (resp TCPData, res TxBroadcastResponse, resFile FileBroadcastResponse) {
	if PackageName == "transaction" && Method == "addTransaction" {

		_, res, _ = SendObject(obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

	} else if PackageName == "file" && Method == "getchunkdata" {

		_, _, resFile = SendObject(obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

	} else if PackageName == "file" && Method == "addchunk" {

		_, _, resFile = SendObject(obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

	} else {

		SendObject(obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)
		m := new(MyMessage)
		m.Value, _ = json.Marshal(*obj)

		host, port, err := net.SplitHostPort(validator.CurrentValidator.ValidatorSoketIP)
		host2, port2, _ := net.SplitHostPort(validatorObj.ValidatorSoketIP)

		ip := net.ParseIP(host)
		m.FromAddr = ip
		x, _ := strconv.Atoi(port)
		m.FromPort = uint16(x)

		ip = net.ParseIP(host2)
		m.Key = validatorObj.ValidatorPublicKey
		x, _ = strconv.Atoi(port2)
		i := uint(x)
		err = sendToClient(m, host2, i)
		fmt.Println(err)
	}
	return resp, res, resFile
}

//BoardcastingTCP Object
func BoardcastingTCP(obj interface{}, Method, PackageName string) (TxBroadcastResponse, FileBroadcastResponse) {
	var res TxBroadcastResponse

	var resFile FileBroadcastResponse
	for _, validatorObj := range validator.ValidatorsLstObj {
		if !validatorObj.ValidatorRemove {
			if validatorObj.ValidatorIP == validator.CurrentValidator.ValidatorIP {
				_, res, resFile = sendToLocalNode(&obj, &validatorObj, PackageName, Method)
			} else {
				_, res, resFile = sendToNode(&obj, &validatorObj, PackageName, Method)
			}
		}
	}
	return res, resFile
}

//SendObject to spacific miner
func SendObject(obj interface{}, Validatorpublickey, Method, PackageName, ValidatorSoketIP string) (TCPData, TxBroadcastResponse, FileBroadcastResponse) {

	var responseObj TxBroadcastResponse
	var responsechunkObj FileBroadcastResponse
	jsonObj, _ := json.Marshal(obj)

	signature := cryptogrpghy.SignPKCS1v15(string(jsonObj), *cryptogrpghy.ParsePEMtoRSAprivateKey(validator.CurrentValidator.ValidatorPrivateKey))

	objTCP := TCPData{jsonObj, validator.CurrentValidator.ValidatorIP, Method, PackageName, signature}

	netObj := NetStruct{}

	hashedkey := cryptogrpghy.CreateSHA1(Validatorpublickey)
	netObj.Encryptedkey, _ = cryptogrpghy.PublicEncrypt(Validatorpublickey, hashedkey)
	byteData, _ := json.Marshal(objTCP)
	strofdata := string(byteData)
	netObj.Encrypteddata = cryptogrpghy.KeyEncrypt(hashedkey, strofdata)

	byteData, _ = json.Marshal(netObj)

	strerr, returnByte := globalPkg.SendBroadCast(byteData, ValidatorSoketIP+"/a021d8007a2c590bc64ff2338d34c4e2", "POST")

	if objTCP.PackageName == "transaction" && Method == "addTransaction" {
		if strerr != "" {
			responseObj.Valid = true
		} else {
			json.Unmarshal(returnByte, &responseObj)
		}
	}

	if objTCP.PackageName == "file" && Method == "getchunkdata" {
		if strerr != "" {
			responsechunkObj.Valid = false
			// responseObj.TxID =
		} else {
			json.Unmarshal(returnByte, &responsechunkObj)
		}

	}

	if objTCP.PackageName == "file" && Method == "addchunk" {
		if strerr != "" {
			responsechunkObj.Valid = false

		} else {
			json.Unmarshal(returnByte, &responsechunkObj)
		}

	}

	return objTCP, responseObj, responsechunkObj
}

func ReadTxResponseData(conn net.Conn, fileName string) TxBroadcastResponse {
	var txResponse TxBroadcastResponse

	bufferFileSize := make([]byte, 10)

	conn.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	newFile, errf := os.Open(fileName)
	if errf != nil {
		fmt.Println(errf)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, conn, (fileSize - receivedBytes))
			conn.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, conn, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}

	file2, _ := os.Open("_" + fileName)
	fileBytes, _ := ioutil.ReadAll(file2)
	defer file2.Close()
	err2 := json.Unmarshal(fileBytes, &txResponse)
	fmt.Println("\n json.Unmarshal(buffer[:n], &txResponse) error: ", err2)
	var txResponse2 TxBroadcastResponse
	mapstructure.Decode(txResponse, &txResponse2) //smart life hack
	fmt.Println("the Response from broadcast handle:", txResponse2)

	// if err1 != nil {
	// 	fmt.Println("broadcastTcp read data error1:", err1)
	// }
	return txResponse2
}

// ReadChunkResponse read chunk response
func ReadChunkResponse(conn net.Conn, fileName string) FileBroadcastResponse {
	var chnkResponse FileBroadcastResponse

	bufferFileSize := make([]byte, 10)

	conn.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	newFile, errf := os.Open(fileName)
	if errf != nil {
		fmt.Println(errf)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, conn, (fileSize - receivedBytes))
			conn.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, conn, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}

	file2, _ := os.Open("_" + fileName)
	fileBytes, _ := ioutil.ReadAll(file2)
	defer file2.Close()

	err2 := json.Unmarshal(fileBytes, &chnkResponse)
	fmt.Println("\n json.Unmarshal(buffer[:n], &txResponse) error: ", err2)

	var chnkRspns2 FileBroadcastResponse
	// fmt.Println("SSSS ", string(chnkRspns2.ChunkData))
	mapstructure.Decode(chnkResponse, &chnkRspns2) //smart life hack
	return chnkRspns2
}

func SendTokenImg(obj string, Validatorpublickey, Method, PackageName, ValidatorSoketIP string) TCPData {
	jsonObj, _ := json.Marshal(obj)

	objTCP := TCPData{jsonObj, validator.CurrentValidator.ValidatorIP, Method, PackageName, ""}

	netObj := NetStruct{}

	netObj.Encryptedkey = "key"
	byteDat, _ := json.Marshal(objTCP)
	strofdata := string(byteDat)
	netObj.Encrypteddata = strofdata //cryptogrpghy.KeyEncrypt(hashedkey, strofdata)

	byteData, _ := json.Marshal(netObj)
	globalPkg.SendBroadCast(byteData, ValidatorSoketIP+"/a021d8007a2c590bc64ff2338d34c4e2", "POST")

	return objTCP
}

func BoardcastingTokenImgUDP(obj string, Method, PackageName string) {
	// if PackageName == "transaction" {
	// }
	for _, validatorObj := range validator.ValidatorsLstObj {
		if !validatorObj.ValidatorRemove {
			if validatorObj.ValidatorIP == validator.CurrentValidator.ValidatorIP {

				SendTokenImg(obj, validatorObj.ValidatorPublicKey, Method, PackageName, validator.CurrentValidator.ValidatorSoketIP)
			} else {

				SendTokenImg(obj, validatorObj.ValidatorPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)
			}
		}
	}
}
func FillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}
