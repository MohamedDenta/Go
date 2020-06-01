package broadcastHandle

import (
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"

	ecc "../ECC"
	"../account"
	"../accountdb"
	"../block"
	"../broadcastTcp"
	"../globalPkg" //alaa
	"../heartbeat"
	"../ledger"
	"../proofofstake"
	"../token"
	"../validator"
)

//BroadcastHandle boardcast to listen
func BroadcastHandle(w http.ResponseWriter, req *http.Request) {

	bufferData := broadcastTcp.NetStruct{}
	DataObj := broadcastTcp.NetStruct{}
	tCPDataObj := broadcastTcp.TCPData{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&bufferData); err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		return
	}

	mapstructure.Decode(bufferData, &DataObj) //smart life hack
	if DataObj.Encryptedkey == "key" {
		jsonData := DataObj.Encrypteddata
		json.Unmarshal([]byte(jsonData), &tCPDataObj)
		for _, obj := range validator.ValidatorsLstObj {
			if obj.ValidatorIP == tCPDataObj.ValidatorIP {
				if !obj.ValidatorRemove {
					if tCPDataObj.PackageName == "addtokenimg" {
						var tokenimgdata string
						json.Unmarshal(tCPDataObj.Obj, &tokenimgdata)
						s := strings.Split(tokenimgdata, "_")
						tokenid := s[0]
						tokendata0 := token.FindTokenByid(tokenid)
						tokendata0.TokenIcon = s[1]
						token.UpdateTokendb(tokendata0)
					}
				}
			}
		}

	}
	if DataObj.Encryptedkey != "key" {
		err := json.Unmarshal([]byte(DataObj.Encrypteddata), &tCPDataObj) // data here is just marshlled
		if err == nil {
			switch tCPDataObj.PackageName {
			case "ledger for new node":
				BoardcastHandleLedgerfornewNode(tCPDataObj)

			case "Send public key back":
				BoardcastHandleSendPKback(tCPDataObj)
			case "Attached public key":
				BoardcastHandleAttachedPK(tCPDataObj)
			default:
				return
			}
		} else {
			priv := validator.CurrentValidator.ECCPrivateKey
			if priv != nil {
				data, err := ecc.Decrypt(priv, DataObj.Encrypteddata)
				if err != nil {
					fmt.Println("Can't decrypt the object plz check ECC private key curve value")
					panic(err)
				}
				jsonData := string(data)
				if jsonData != "" {
					json.Unmarshal([]byte(jsonData), &tCPDataObj)
					if tCPDataObj.PackageName == "ledger" && len(accountdb.GetAllAccounts()) == 0 {
						var ledgObj ledger.Ledger
						json.Unmarshal(tCPDataObj.Obj, &ledgObj)
						ledger.PostLedger(ledgObj)
					} else {

						for _, obj := range validator.ValidatorsLstObj {

							if obj.ValidatorIP == tCPDataObj.ValidatorIP {
								if !obj.ValidatorRemove {

									if obj.ECCPublicKey.Curve == nil {
										fmt.Println("Failed at Verfication process with ECC public key : ", obj.ECCPublicKey)
										return
									}
									verifyStatus := ecc.Verification(obj.ECCPublicKey, tCPDataObj.Signature.SignHash, tCPDataObj.Signature.R, tCPDataObj.Signature.S)
									if verifyStatus == false {
										verificatioError := errors.New("signature: failed to vrifiy your signature")
										panic(verificatioError)
									}
									switch tCPDataObj.PackageName {
									case "account":
										account.BoardcastHandleAccount(tCPDataObj)
									case "account module":
										account.BoardcastHandleAccountModule(tCPDataObj)

									case "transaction":
										BoardcastHandleTransaction(tCPDataObj, w)
									case "block":
										block.BoardcatHandleBlock(tCPDataObj)

									case "heartBeat":
										heartbeat.BoardcastHandleHeartbeat(tCPDataObj)
									case "proofOfStake":
										proofofstake.BoardcastHandleProofStak(tCPDataObj)

									case "admin":
										BoardcastHandleAdmin(tCPDataObj)

									case "token":
										BoardcastHandleToken(tCPDataObj)
									case "Delete Session":
										account.BoardcastHandleDeleteSession(tCPDataObj)
									case "Add Session":
										account.BoardcastHandleAddSession(tCPDataObj)
									case "create ownership": // for create or update ownership
										account.BoardcastHandleOwnership(tCPDataObj)
									case "validator":
										BoardcastHandleValidator(tCPDataObj)

									case "confirmedvalidator":
										BoardcastHandleConfirmValidator(tCPDataObj)

									case "Add Service":
										BoardcastHandleService(tCPDataObj)
									case "AddAndUpdateLog":
										BoardcastHandleLog(tCPDataObj)

									case "savepk":
										account.BoardcastHandleSavePK(tCPDataObj)

									case "file":
										BoardcastHandleFile(tCPDataObj, w)
									default:
										return
									}

								}
							}
						}
					}

				}
			} else {
				fmt.Println("ECC Private key is nil, can't decrypt")
			}
		}

	}
}
