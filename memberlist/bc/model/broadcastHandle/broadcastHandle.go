package broadcastHandle

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"../transactionValidation"
	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"

	//	"log"
	"strings"

	"../service"

	"../account"
	"../admin"
	"../cryptogrpghy"
	"../ledger"

	"../accountdb"
	"../block"
	"../broadcastTcp"
	"../token"
	"../tokenModule"

	// "../errorpk" // set the waiting time

	//"time"

	file "../filestorage"
	"../globalPkg"
	"../heartbeat"
	"../proofofstake"
	"../transaction"
	"../validator"
	"github.com/mitchellh/mapstructure"
)

type WriteCounter struct {
	Total uint64
}

const BUFFERSIZE = 1024

func delaySecond(n time.Duration) {
	time.Sleep(n * time.Second)
}

func (WCount *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	WCount.Total += uint64(n)
	WCount.PrintProgress()
	return n, nil
}

func (WCount WriteCounter) PrintProgress() {

	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	//print bytes in a meaningful way
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(WCount.Total))
}

func DownloadFile(filepath string, url string) error {

	//download file .tmp file and remove .tmp extension when finnished
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

	// print new line when downloading finnished
	fmt.Print("\n")

	//err = os.Rename(name+".tmp", filepath)
	//if err != nil {
	//	return err
	//}

	return nil
}

// func ListenConnection(conn net.Conn) {
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
						fmt.Println("tCPDataObj.data", tokenimgdata)
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
		hashedkey, _ := cryptogrpghy.Decrypt(validator.CurrentValidator.ValidatorPublicKey, validator.CurrentValidator.ValidatorPrivateKey, DataObj.Encryptedkey)
		if hashedkey != "" {
			jsonData := cryptogrpghy.KeyDecrypt(hashedkey, DataObj.Encrypteddata)
			if jsonData != "" {
				json.Unmarshal([]byte(jsonData), &tCPDataObj)
				if tCPDataObj.PackageName == "ledger" && len(accountdb.GetAllAccounts()) == 0 {

					var ledgObj ledger.Ledger
					json.Unmarshal(tCPDataObj.Obj, &ledgObj)
					ledger.PostLedger(ledgObj)
					// globalPkg.IsDown = false

				} else {
					for _, obj := range validator.ValidatorsLstObj {
						//	fmt.Println(obj.ValidatorIP, tCPDataObj.Validatorpublickey)
						if obj.ValidatorIP == tCPDataObj.ValidatorIP {
							if !obj.ValidatorRemove {
								// if cryptogrpghy.VerifyPKCS1v15(tCPDataObj.Signature, string(tCPDataObj.Obj), *cryptogrpghy.ParsePEMtoRSApublicKey(obj.ValidatorPublicKey)) {
								switch tCPDataObj.PackageName {
								case "account":
									var accountObjc accountdb.AccountStruct

									json.Unmarshal(tCPDataObj.Obj, &accountObjc)
									fmt.Println(accountObjc)
									// var l []string
									// l = tCPDataObj.CurrentTime
									// accountObjc.AccountLastUpdatedTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", l[0])
									if tCPDataObj.Method == "POST" {
										account.AddAccount(accountObjc)
										lst := account.GetUserObjLst()
										for index, data := range lst {
											if data.Account.AccountName == accountObjc.AccountName {
												account.RemoveUserFromtemp(index)
												break
											}
										}
									} else if tCPDataObj.Method == "PUT" {
										fmt.Println("+++++++++++++++++++", accountObjc)
										account.UpdateAccount(accountObjc)
										lst := account.GetUserObjLst()
										for index, data := range lst {
											if data.Account.AccountName == accountObjc.AccountName {
												fmt.Println("///////update")
												account.RemoveUserFromtemp(index)
												break
											}
										}
									} else if tCPDataObj.Method == "set public key" {
										account.SetPublicKey(accountObjc)
									} else if tCPDataObj.Method == "Resetpass" {
										account.UpdateAccount2(accountObjc)
										lst2 := account.GetResetPasswordData()
										for index, data := range lst2 {
											fmt.Println("---------9999666333")
											if data.Email == accountObjc.AccountEmail {

												account.RemoveResetpassFromtemp(index)
												break
											}
										}

									} else if tCPDataObj.Method == "update2" { /////change status
										account.UpdateAccount2(accountObjc)

									}

								case "account module":
									var accmodObjec account.ResetPasswordData
									var accmodObjecuser account.User
									//	accmodObjecuser.CurrentTime , _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
									if tCPDataObj.Method == "addRestPassword" {
										// mapstructure.Decode(tCPDataObj.Obj, &accmodObjec)
										json.Unmarshal(tCPDataObj.Obj, &accmodObjec)
										fmt.Println("your object : ", accmodObjec)
										// var l []string
										// l = tCPDataObj.CurrentTime
										// u := l[0]
										// accmodObjec.CurrentTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", u)

										// israa to not repeat reset password codes
										lst := account.GetResetPasswordData()
										add := true
										if len(lst) != 0 {
											for index, data := range lst {
												if data.Email == accmodObjec.Email {
													// account.UpdateResetpassObjInTemp(index, accmodObjec)
													account.UpdateResetpassObjInTemp(index, accmodObjec)
													add = false
													break
												}
											}
										}
										if add == true {
											account.AddResetpassObjInTemp(accmodObjec)
										}
										//end update

									} else if tCPDataObj.Method == "adduser" {
										// mapstructure.Decode(tCPDataObj.Obj, &accmodObjecuser)
										json.Unmarshal(tCPDataObj.Obj, &accmodObjecuser)
										fmt.Println("your object : ", accmodObjecuser)

										// var l []string
										// l = tCPDataObj.CurrentTime
										// fmt.Println(" my list")
										// fmt.Println(l[0])
										// u := l[0]
										// accmodObjecuser.CurrentTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", u)
										// fmt.Println("accmodObjecuser.CurrentTime : ++++++++++")
										// fmt.Println(accmodObjecuser.CurrentTime)
										account.AddUserIntemp(accmodObjecuser)
										// // 	// israa to not repeat update codes
										// lst := account.GetUserObjLst()
										// add := true
										// if len(lst) != 0 {
										// 	for index, data := range lst {
										// 		if data.Name == accmodObjecuser.Name {
										// 			account.UpdateUserTotemp(index, accmodObjecuser)
										// 			add = false
										// 			break
										// 		}
										// 	}
										// }
										// if add == true {
										// 	account.AddUserIntemp(accmodObjecuser)
										// }
										// end update
									} else if tCPDataObj.Method == "Update" {
										json.Unmarshal(tCPDataObj.Obj, &accmodObjecuser)
										account.UpdateconfirmAtribute(accmodObjecuser)
									}

								case "transaction":
									if tCPDataObj.Method == "addTransaction" {
										//var transobjec transaction.Transaction
										var txMix transaction.MixedTxStruct
										// lst := tCPDataObj.CurrentTime

										// transobjec.TransactionTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
										// mapstructure.Decode(tCPDataObj.Obj, &transobjec)
										json.Unmarshal(tCPDataObj.Obj, &txMix)

										responseData := broadcastTcp.TxBroadcastResponse{
											TxID: txMix.TxObj.TransactionID, Valid: true,
										}

										//transobjec.TransactionTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", lst[0])
										if txValid := transactionValidation.ValidateTx2(txMix.DigitalTxObj, txMix.TxObj); txValid == "true" {
											fmt.Println("--------------*-", txValid)

											transaction.AddTransaction(txMix.TxObj)
											for _, validatorObj := range validator.ValidatorsLstObj {
												// if validatorObj.ValidatorPublicKey == tCPDataObj.ValidatorPublicKey {
												if validatorObj.ValidatorIP == tCPDataObj.ValidatorIP {
													validatorObj.ValidatorStakeCoins = validatorObj.ValidatorStakeCoins + globalPkg.GlobalObj.TransactionStakeCoins
													validator.UpdateValidator(validatorObj)
													break
												}
											}
										} else {
											fmt.Println("\n --------------*- broadcast handle transaction validation:", txValid)
											responseData.Valid = false
										}
										//var sizewrite int
										responseByteData, _ := json.Marshal(responseData)
										globalPkg.SendResponse(w, responseByteData)
										return
										// str := "ino" + tim + "_" + hashpk[0:16] + "_" + PackageName + ".txt"
										// ioutil.WriteFile("_"+fileName, responseByteData, 0644)

										// // file, err1 := os.Open(path + "/" + str)
										// file, err1 := os.Open("_" + fileName)
										// defer file.Close()
										// if err1 != nil {
										// 	fmt.Println("error2", err)
										// 	// return objTCP
										// }
										// fileInfo, err := file.Stat()
										// if err != nil {
										// 	fmt.Println(err)
										// 	// return objTCP
										// }

										// fileSize := broadcastTcp.FillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
										// conn.Write([]byte(fileSize))

										// // fmt.Println("Start sending file!", fileName, " ", fileSize)
										// sendBuffer := make([]byte, BUFFERSIZE)
										// for {
										// 	_, err = file.Read(sendBuffer)
										// 	if err == io.EOF {
										// 		break
										// 	}
										// 	conn.Write(sendBuffer)
										// }

										// conn.Write([]byte(broadcastTcp.FillString(strconv.Itoa(len(responseByteData)), 64)))
										// sendBuffer := make([]byte, broadcastTcp.BUFFERSIZE)
										// for index := 0; index <= len(responseByteData); index = index + broadcastTcp.BUFFERSIZE {
										// 	if index+broadcastTcp.BUFFERSIZE >= len(responseByteData) {
										// 		sendBuffer = responseByteData[index:len(responseByteData)]
										// 	} else {
										// 		sendBuffer = responseByteData[index : index+broadcastTcp.BUFFERSIZE]
										// 	}
										// 	conn.Write(sendBuffer)
										// }
										// nOfBytes, err := conn.WriteToUDP(responseByteData, address)
										// if len(string(byteData))%256 == 0 {
										// 	fmt.Fprintf(conn, string(byteData)+" ")
										// 	sizewrite = len(string(byteData)) + 1
										// } else {
										// 	fmt.Fprintf(conn, string(byteData))
										// 	sizewrite = len(string(byteData))
										// }
										//fmt.Fprintf(conn, string(responseByteData))
										// fmt.Println("\n transaction response conn.WriteToUDP error/N of bytes :", err, "/", nOfBytes)

										//senderNodeValidator := validator.FindValidatorByValidatorIP(tCPDataObj.ValidatorIP)
										// broadcast the transaction validation response
										//broadcastTcp.SendObject(responseData, senderNodeValidator.ValidatorPublicKey, "response", "transaction", senderNodeValidator.ValidatorSoketIP)

									} else if tCPDataObj.Method == "addTokenTransaction" {
										var tokenTx transaction.Transaction

										json.Unmarshal(tCPDataObj.Obj, &tokenTx)

										transaction.AddTransaction(tokenTx)

									}

								case "block":
									var blockObjec block.BlockStruct
									// lst := tCPDataObj.CurrentTime
									// blockObjec.BlockTimeStamp, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
									// mapstructure.Decode(tCPDataObj.Obj, &blockObjec)
									json.Unmarshal(tCPDataObj.Obj, &blockObjec)

									var blockObjec2 block.BlockStruct
									blockObjec2 = blockObjec
									blockObjec2.BlockTransactions = nil
									fmt.Println("obj.TransactionTime**************")
									for _, obj := range blockObjec.BlockTransactions {
										// obj.TransactionTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", lst[index])
										// fmt.Println("obj.TransactionTime this is fixed**************", lst[index], obj.TransactionTime)
										blockObjec2.BlockTransactions = append(blockObjec2.BlockTransactions, obj)
									}
									fmt.Println("obj.TransactionTime**************")
									// blockObjec2.BlockTimeStamp, _ = time.Parse("2006-01-02 03:04:05 PM -0000", lst[len(lst)-1])
									//	blockObjec.BlockTimeStamp ,_ = time.Parse("2006-01-02 03:04:05 PM -0000", lst[0])
									fmt.Println("your block sis : ", blockObjec2)
									block.AddBlock(blockObjec2, false)

								case "heartBeat":
									var message heartbeat.Message
									var heartbeatObjec heartbeat.MinersInfo
									// fmt.Println("the local server is := ", conn.LocalAddr())
									// fmt.Println("the other servers server are := ", conn.RemoteAddr())
									message.TimeStamp = globalPkg.UTCtime()
									// mapstructure.Decode(tCPDataObj.Obj, &message)
									json.Unmarshal(tCPDataObj.Obj, &message)
									fmt.Println("------*--------------", message)
									if message.UpdateExist == false {

										heartbeatObjec.MinerStatus = true
										heartbeatObjec.Message = message
										fmt.Println("++++++++++++++++", message)
										for _, validatorObj := range validator.ValidatorsLstObj {
											if validatorObj.ValidatorIP == heartbeatObjec.Message.MinerIP {
												heartbeat.CompareMinerStatus(heartbeatObjec, validatorObj)
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
											err = DownloadFile("", message.UpdateUrl) // update file path

											delaySecond(30)

											if err != nil {
												// ...
											} else {
												// ...
											}
											if err != nil {
												panic(err)
											}
											fmt.Println("Download Complete")

											if err != nil {
												log.Println(err)
											}

											fmt.Println("Running command and waiting for it to finish...")
											cmd := exec.Command("sudo", "systemctl", "start", "auto.service")
											fmt.Println("killing old build ")
											cmd = exec.Command("pkill", "main")
											cmd = exec.Command("sudo", "systemctl", "stop", "auto.service")
											cmd.Stderr = os.Stderr
											cmd.Stdin = os.Stdin

											out, err := cmd.Output()
											if err != nil {
												fmt.Println("Err", err)
											} else {
												fmt.Println("OUT:", string(out))
											}
											err = cmd.Run()
											fmt.Println("we can't close build because", err)

											// get update version and update url and write on file
											// fmt.Println("**Update Version **", message)
											// fo, err := os.Create("update.txt")
											// if err != nil {
											// 	fmt.Println(err)
											// }
											// if _, err := fo.Write([]byte("update version: ")); err != nil {
											// 	fmt.Println(err)
											// }
											// x,err:=json.Marshal(message.UpdateVersion)
											// if _, err := fo.Write(x); err != nil {
											// 	fmt.Println(err)
											// }
											// time.Sleep(5*time.Second)
											// if _, err := fo.Write([]byte("\r\n")); err != nil {
											// 	fmt.Println(err)
											// }
											// message.UpdateUrl="update url : "+message.UpdateUrl
											// if _, err := fo.Write([]byte(message.UpdateUrl)); err != nil {
											// 	fmt.Println(err)
											// }
										}
									}

								case "proofOfStake":
									var proofOSObjec proofofstake.WinningValidatorStruct
									proofOSObjec.TimeStamp = globalPkg.UTCtime()
									proofOSObjec.WinnerValidator.ValidatorRegisterTime = globalPkg.UTCtime()
									proofOSObjec.CurrentNode.ValidatorRegisterTime = globalPkg.UTCtime()
									// mapstructure.Decode(tCPDataObj.Obj, &proofOSObjec)
									json.Unmarshal(tCPDataObj.Obj, &proofOSObjec)
									proofofstake.ForgeTheBlock(proofOSObjec)

								case "admin":
									var AdminObj admin.AdminStruct
									if tCPDataObj.Method == "addadmin" {
										// mapstructure.Decode(tCPDataObj.Obj, &AdminObj)
										json.Unmarshal(tCPDataObj.Obj, &AdminObj)
										fmt.Println("AdminObj", AdminObj)
										admin.CreateAdmin(AdminObj)
									} else if tCPDataObj.Method == "updateadmin" {
										// mapstructure.Decode(tCPDataObj.Obj, &AdminObj)
										json.Unmarshal(tCPDataObj.Obj, &AdminObj)
										fmt.Println("AdminObj", AdminObj)
										admin.UpdateAdmindb(AdminObj)
									}

								case "token":
									fmt.Println(" ***********   Mohamed   ***********************")
									var TokenObj token.StructToken
									// mapstructure.Decode(tCPDataObj.Obj, &AdminObj)

									if tCPDataObj.Method == "addtoken" {
										json.Unmarshal(tCPDataObj.Obj, &TokenObj)
										fmt.Println(TokenObj)
										fmt.Println("TokenObj")
										tokenModule.AddToken(TokenObj)
									} else if tCPDataObj.Method == "updatetoken" {
										json.Unmarshal(tCPDataObj.Obj, &TokenObj)
										fmt.Println("TokenObj", TokenObj)
										token.UpdateTokendb(TokenObj)
									}
								case "Delete Session":
									var sessionid accountdb.AccountSessionStruct
									json.Unmarshal(tCPDataObj.Obj, &sessionid)
									account.RemoveSessionFromtemp(sessionid)
								case "Add Session":
									var sessionid accountdb.AccountSessionStruct
									json.Unmarshal(tCPDataObj.Obj, &sessionid)
									account.AddSessioninTemp(sessionid)
								case "validator":
									if tCPDataObj.Method == "POST" {
										var timpValidator validator.TempValidator
										json.Unmarshal(tCPDataObj.Obj, &timpValidator)
										validator.AddValidatorTemporary(timpValidator)
									} else if tCPDataObj.Method == "PUT" {
										var validatorObj validator.ValidatorStruct
										json.Unmarshal(tCPDataObj.Obj, &validatorObj)
										validator.UpdateValidator(validatorObj)
									}

								case "confirmedvalidator":
									var validatorObj validator.ValidatorStruct
									json.Unmarshal(tCPDataObj.Obj, &validatorObj) //add the validator in validators list after admin confirmation
									validator.AddValidator(validatorObj)
								case "Add Service":
									if tCPDataObj.Method == "Tmp" {
										var serviceobj service.ServiceStruct
										json.Unmarshal(tCPDataObj.Obj, &serviceobj)
										fmt.Println("44444444444")
										service.AddserviceInTmp(serviceobj)
									}
									if tCPDataObj.Method == "DB" {
										var serviceobj service.ServiceStruct

										json.Unmarshal(tCPDataObj.Obj, &serviceobj)

										service.AddAndUpdateServiceObj(serviceobj)
										servicetemp := service.GetAllservice()
										for index, obj := range servicetemp {
											if serviceobj.PublicKey == obj.PublicKey && serviceobj.ID == obj.ID {
												service.RemoveServicefromTmp(index)
												break
											}

										}
									}
								// case "AddAndUpdateLog":
								// 	var logobj logpkg.LogStruct
								// 	json.Unmarshal(tCPDataObj.Obj, &logobj)

								// 	logfunc.WriteAndUpdateLog(logobj)

								case "savepk":
									fmt.Println("broadcast savepk ")
									var savepkobj account.SavePKStruct
									json.Unmarshal(tCPDataObj.Obj, &savepkobj)
									account.SavePKAddress(savepkobj)

								case "file":
									if tCPDataObj.Method == "addchunk" {
										var chunkobj file.Chunkdb
										json.Unmarshal(tCPDataObj.Obj, &chunkobj)
										var responseDataChunk broadcastTcp.FileBroadcastResponse
										if file.AddChunk(chunkobj) {
											responseDataChunk.Valid = true
										}
										responseByteData, _ := json.Marshal(responseDataChunk)
										globalPkg.SendResponse(w, responseByteData)
										return
									} else if tCPDataObj.Method == "addfile" {
										var fileobj transaction.Transaction
										json.Unmarshal(tCPDataObj.Obj, &fileobj.Filestruct)
										transaction.AddTransaction(fileobj)
									} else if tCPDataObj.Method == "deletefile" {
										var fileobj transaction.Transaction
										json.Unmarshal(tCPDataObj.Obj, &fileobj.Filestruct)
										transaction.AddTransaction(fileobj)
									} else if tCPDataObj.Method == "getchunkdata" {
										var chnkObj file.Chunkdb
										json.Unmarshal(tCPDataObj.Obj, &chnkObj)
										retrievedObj := file.FindChunkByid(chnkObj.Chunkid)
										var responseDataChunk broadcastTcp.FileBroadcastResponse
										if retrievedObj.Chunkhash == chnkObj.Chunkhash {
											if len(retrievedObj.Chunkdata) != 0 {
												responseDataChunk.Valid = true
												responseDataChunk.ChunkData = retrievedObj.Chunkdata
												responseByteData, _ := json.Marshal(responseDataChunk)
												globalPkg.SendResponse(w, responseByteData)
												return
											}
										}
									} else if tCPDataObj.Method == "sharefile" {
										var sharefileobj file.SharedFile
										json.Unmarshal(tCPDataObj.Obj, &sharefileobj)
										file.AddSharedFile(sharefileobj)
									} else if tCPDataObj.Method == "updatesharefile" {
										var sharefileobj file.SharedFile
										json.Unmarshal(tCPDataObj.Obj, &sharefileobj)
										file.Updatesharefile(sharefileobj)
									} else if tCPDataObj.Method == "deleteaccountindex" {
										var sharefileobj file.SharedFile
										json.Unmarshal(tCPDataObj.Obj, &sharefileobj)
										file.DeleteSharedFile(sharefileobj.AccountIndex)
									} else if tCPDataObj.Method == "updateaccountFilelist" {
										var accountObj accountdb.AccountStruct
										json.Unmarshal(tCPDataObj.Obj, &accountObj)
										account.UpdateAccount2(accountObj)
									}

								default:
									return
								}
								// }
							}
						}
					}
				}
				//	retrieveDataFromTemp()
			}
		}
	}
	// }
}
