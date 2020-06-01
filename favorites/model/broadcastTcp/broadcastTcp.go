package broadcastTcp

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"../globalPkg"

	ecc "../ECC"
	"../validator"
)

//TempData contain array tcpdata
var TempData []TCPData

//TCPData struct contain data about object,package name ,method
type TCPData struct {
	Obj         []byte
	ValidatorIP string
	Method      string
	PackageName string
	Signature   SignatureObj
}

//SignatureObj containing all parametares for signing and verification
type SignatureObj struct {
	Signature []byte
	R         *big.Int
	S         *big.Int
	SignHash  []byte
}

// NetStruct contain key , data encrypt
type NetStruct struct {
	Encryptedkey  string
	Encrypteddata []byte
}

//TxBroadcastResponse return reponse of transaction valid or not and transaction id
type TxBroadcastResponse struct {
	TxID  string
	Valid bool
}

//ManageObjectTime to manage send object time
type ManageObjectTime struct {
	TCPObj             TCPData
	ValidatorSocket    string
	Validatorpublickey ecdsa.PublicKey
	Transactionid      string
}

//ArrManageObject array append manage send object
var ArrManageObject []ManageObjectTime

//TransactionReponse transaction reponse count and status for transaction
type TransactionReponse struct {
	Count  int
	Status bool
}

//MapReponse to store reponse of transaction
var MapReponse = make(map[string]TransactionReponse)

//Rwm mutex to lock and unlock map
var Rwm sync.RWMutex

//FileBroadcastResponse return reponse of file that upload chunk in that server or not
type FileBroadcastResponse struct {
	ChunkData []byte
	Valid     bool
}

//TempNotRecieving IS store tcpdata an ip
// type TempNotRecieving struct {
// 	TCPData
// 	ValidatorSoketIP string
// }

// var temp []TempNotRecieving

//BoardcastingTCP Object
func BoardcastingTCP(obj interface{}, Method, PackageName string) (TxBroadcastResponse, FileBroadcastResponse) {
	var res TxBroadcastResponse
	var resFile FileBroadcastResponse

	for _, validatorObj := range validator.ValidatorsLstObj {
		if !validatorObj.ValidatorRemove {
			if validatorObj.ValidatorIP == validator.CurrentValidator.ValidatorIP {

				if PackageName == "transaction" && Method == "addTransaction" {
					_, res, _ = SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validator.CurrentValidator.ValidatorSoketIP)

				} else if PackageName == "file" && Method == "getchunkdata" {
					_, _, resFile = SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

				} else if PackageName == "file" && Method == "addchunk" {
					_, _, resFile = SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)
				} else {
					SendObject(obj, validator.CurrentValidator.ECCPublicKey, Method, PackageName, validator.CurrentValidator.ValidatorSoketIP)
				}
			} else {
				if PackageName == "transaction" && Method == "addTransaction" {
					_, res, _ = SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

				} else if PackageName == "file" && Method == "getchunkdata" {
					_, _, resFile = SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)

				} else if PackageName == "file" && Method == "addchunk" {
					_, _, resFile = SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)
				} else {
					SendObject(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)
				}
			}
		}
	}

	return res, resFile
}

//SendObject to spacific miner
func SendObject(obj interface{}, Validatorpublickey ecdsa.PublicKey, Method, PackageName, ValidatorSoketIP string) (TCPData, TxBroadcastResponse, FileBroadcastResponse) {
	jsonObj, _ := json.Marshal(obj)
	var txid string
	var responseObj TxBroadcastResponse
	var responsechunkObj FileBroadcastResponse
	// split method name add transaction and transaction id
	if PackageName == "transaction" && strings.Contains(Method, "addTransaction") {
		methodName := strings.Split(Method, "+")
		Method = "addTransaction"
		txid = methodName[1]
	}

	var Sig SignatureObj
	Sig.Signature, Sig.R, Sig.S, Sig.SignHash = ecc.Sign(string(jsonObj), validator.CurrentValidator.ECCPrivateKey)

	objTCP := TCPData{jsonObj, validator.CurrentValidator.ValidatorIP, Method, PackageName, Sig}

	transactionReponseObj := TransactionReponse{}

	netObj := NetStruct{}

	byteData, _ := json.Marshal(objTCP)

	if PackageName != "Attached public key" && PackageName != "Send public key back" && PackageName != "ledger for new node" {
		if Validatorpublickey.Curve == nil {
			fmt.Println("empty Public key in SendObject(): cann't encrypt")
		}
		enc, err := ecc.Encrypt(&Validatorpublickey, byteData)
		if err != nil {
			fmt.Println("err in encrypt :    ", err)
		}
		netObj.Encrypteddata = enc
	} else {
		netObj.Encrypteddata = byteData //sending the data without encryption for new nodes
	}

	byteData, _ = json.Marshal(netObj)
	strerr, returnByte := globalPkg.SendBroadCast(byteData, ValidatorSoketIP+"/a021d8007a2c590bc64ff2338d34c4e2", "POST")

	// check package name transaction
	if PackageName == "transaction" && Method == "addTransaction" {
		if strerr != "" {
			responseObj.Valid = true
			countTransaction(txid)
		} else {

			json.Unmarshal(returnByte, &responseObj)

			if _, ok := MapReponse[responseObj.TxID]; ok {
				//if trans id exist in map reponse increase count
				transactionReponseObj = GetMap(responseObj.TxID)
				transactionReponseObj.Count = transactionReponseObj.Count + 1
				transactionReponseObj.Status = GetMap(responseObj.TxID).Status
			} else {
				transactionReponseObj.Count = 1
				transactionReponseObj.Status = true
			}
			if responseObj.Valid == false {
				transactionReponseObj.Status = false
			}
			SetMap(responseObj.TxID, transactionReponseObj) // set values of count and status.
		}

	}

	if objTCP.PackageName == "file" && Method == "getchunkdata" {
		if strerr != "" {
			responsechunkObj.Valid = false
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

//SetMap set map with count and status of transaction
func SetMap(key string, value TransactionReponse) {
	// Rwm.Lock()
	// defer Rwm.Unlock()
	// MapReponse[key] = value
	Rwm.Lock()
	MapReponse[key] = value
	Rwm.Unlock()
}

//GetMap get transaction reponse from map
func GetMap(key string) TransactionReponse {
	// Rwm.RLock()
	// defer Rwm.RUnlock()
	// return MapReponse[key]
	Rwm.RLock()
	value := MapReponse[key]
	Rwm.RUnlock()
	return value

}

//countTransaction if read size = 0 to count to be equal sum of not remove validator  --not active validator
func countTransaction(TXid string) {
	transactionReponseObj := TransactionReponse{}
	transactionReponseObj = GetMap(TXid)
	if transactionReponseObj.Count != 0 { //check not empty map that txid exist in map increase count
		transactionReponseObj.Count = transactionReponseObj.Count + 1
	} else {
		transactionReponseObj.Count = 1
	}
	SetMap(TXid, transactionReponseObj)
}

// ReadTxResponseData read transaction reponse data
func ReadTxResponseData(conn net.Conn) TxBroadcastResponse {
	var txResponse TxBroadcastResponse

	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	n, err1 := conn.Read(buffer)
	err2 := json.Unmarshal(buffer[:n], &txResponse)
	fmt.Println("\n json.Unmarshal(buffer[:n], &txResponse) error: ", err2)

	if err1 != nil {
		fmt.Println("broadcastTcp read data error1:", err1)
	}
	return txResponse
}

//SendTokenImg send token image
func SendTokenImg(obj string, Validatorpublickey ecdsa.PublicKey, Method, PackageName, ValidatorSoketIP string) TCPData {
	jsonObj, _ := json.Marshal(obj)

	var sig SignatureObj //ask omar omar if u need to fill this sig
	objTCP := TCPData{jsonObj, validator.CurrentValidator.ValidatorIP, Method, PackageName, sig}

	netObj := NetStruct{}

	netObj.Encryptedkey = "key"
	byteDat, _ := json.Marshal(objTCP)
	netObj.Encrypteddata = byteDat //cryptogrpghy.KeyEncrypt(hashedkey, strofdata)

	byteData, _ := json.Marshal(netObj)
	globalPkg.SendBroadCast(byteData, ValidatorSoketIP+"/a021d8007a2c590bc64ff2338d34c4e2", "POST")

	return objTCP
}

//BoardcastingTokenImgUDP boardcast image udp
func BoardcastingTokenImgUDP(obj string, Method, PackageName string) {

	for _, validatorObj := range validator.ValidatorsLstObj {
		if !validatorObj.ValidatorRemove {
			if validatorObj.ValidatorIP == validator.CurrentValidator.ValidatorIP {

				SendTokenImg(obj, validatorObj.ECCPublicKey, Method, PackageName, validator.CurrentValidator.ValidatorSoketIP)
			} else {

				SendTokenImg(obj, validatorObj.ECCPublicKey, Method, PackageName, validatorObj.ValidatorSoketIP)
			}
		}
	}
}
