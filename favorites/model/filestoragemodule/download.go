package filestoragemodule

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"../block"
	"../cryptogrpghy"
	"../filestorage"

	"../account"
	"../accountdb"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../validator"
)

// RetrieveFile retrive file for some user
func RetrieveFile(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "RetrieveFile", "file", "_", "_", "_", 0}

	var obj RetrieveBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&obj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	// check for pk
	acc := account.GetAccountByAccountPubicKey(obj.Publickey)
	if acc.AccountPublicKey != obj.Publickey {
		globalPkg.SendError(w, "error in public key")
		globalPkg.WriteLog(logobj, "error in public key", "failed")
		return
	}
	// check for pwd
	if acc.AccountPassword != obj.Password {
		globalPkg.SendError(w, "error in password")
		globalPkg.WriteLog(logobj, "error in password", "failed")
		return
	}
	// TODO check time
	// Validate Signture
	validSig := false
	pk := account.FindpkByAddress(acc.AccountPublicKey).Publickey
	if pk != "" {
		publickey := cryptogrpghy.ParsePEMtoRSApublicKey(pk)
		// signatureData := obj.FileID + obj.Publickey + obj.Password +
		// 	obj.Time
		signatureData := obj.Publickey + obj.Password + obj.FileID

		validSig = cryptogrpghy.VerifyPKCS1v15(obj.Signture, signatureData, *publickey)
		validSig = true
	} else {
		validSig = false
	}
	if !validSig {
		globalPkg.SendError(w, "you are not allowed to download")
		globalPkg.WriteLog(logobj, "you are not allowed to download", "failed")
		return
	}
	// check is user own this file ?
	files := acc.Filelist
	found := false
	var selectedFile accountdb.FileList
	for _, file := range files {
		if file.Fileid == obj.FileID {
			found = true
			selectedFile = file
			break
		}
	}
	// check if this file share to this account== who take share file can download it
	sharefiles := filestorage.FindSharedfileByAccountIndex(acc.AccountIndex)
	if len(sharefiles.OwnerSharefile) != 0 {
		for _, sharefileobj := range sharefiles.OwnerSharefile {
			if containsfileid(sharefileobj.Fileid, obj.FileID) {
				found = true
				accuntObj := account.GetAccountByAccountPubicKey(sharefileobj.OwnerPublicKey)
				for _, filelistObj := range accuntObj.Filelist {
					if filelistObj.Fileid == obj.FileID {
						selectedFile = filelistObj
						break
					}
				}
			}
		}
	}
	// fmt.Println("selectedFile.FileName ", selectedFile.FileName)
	if !found {
		globalPkg.SendError(w, "You don't have this file or file shared to you")
		globalPkg.WriteLog(logobj, "You don't have this file or file shared to you", "failed")
		return
	}

	// collect file and save it in a temp file
	decryptIndexBlock1 := cryptogrpghy.KeyDecrypt("123456789", selectedFile.Blockindex)
	fmt.Println(" *********** block index ", decryptIndexBlock1)
	blkObj := block.GetBlockInfoByID(decryptIndexBlock1)
	var fStrct filestorage.FileStruct
	for _, tx := range blkObj.BlockTransactions {
		fStrct = tx.Filestruct
		if fStrct.Fileid == selectedFile.Fileid {
			fStrct = tx.Filestruct
			break
		}
	}
	// check active validators
	var actives []validator.ValidatorStruct
	validatorLst := validator.GetAllValidatorsDecrypted()
	for _, valdtr := range validatorLst {
		if valdtr.ValidatorActive {
			actives = append(actives, valdtr)
		}
	}
	var chnkObj filestorage.Chunkdb
	newPath := filepath.Join(uploadPath, fStrct.Fileid+fStrct.FileType)
	file, er := os.OpenFile(newPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0777)
	if er != nil {
		fmt.Println("error in open file ", err)
		globalPkg.SendError(w, "server is down !")
		globalPkg.WriteLog(logobj, "can not oprn file ", "failed")
		return
	}
	defer file.Close()
	notvalidchnkdata := false
	countnotvalidchnkdata := 0
	var res broadcastTcp.FileBroadcastResponse
	var chunkcount int = 0
	for key, value := range fStrct.Mapping {
		for _, chunkDta := range value {
			// time.Sleep(time.Millisecond * 10)
			indofvalidator := contains(actives, chunkDta.ValidatorIP)
			if indofvalidator != -1 {
				validatorObj2 := actives[indofvalidator]
				chnkObj.Chunkid = key
				chnkObj.Chunkhash = chunkDta.Chunkhash
				// _, _, res := broadcastTcp.SendObject(chnkObj, actives[i].ValidatorPublicKey, "getchunkdata", "file", actives[i].ValidatorSoketIP)
				// if validatorObj2.ValidatorPublicKey == validator.CurrentValidator.ValidatorPublicKey {
				// 	_, _, res = broadcastTcp.SendObject(chnkObj, validator.CurrentValidator.ValidatorPublicKey, "getchunkdata", "file", validator.CurrentValidator.ValidatorSoketIP)

				// } else {
				// 	_, _, res = broadcastTcp.SendObject(chnkObj, validatorObj2.ValidatorPublicKey, "getchunkdata", "file", validatorObj2.ValidatorSoketIP)
				// }

				_, _, res = broadcastTcp.SendObject(chnkObj, validatorObj2.ECCPublicKey, "getchunkdata", "file", validatorObj2.ValidatorSoketIP)

				if !res.Valid {
					fmt.Println("server is down")
					notvalidchnkdata = true
					continue
				} else {
					reshashchunk := globalPkg.GetHash(res.ChunkData)
					if reshashchunk != chnkObj.Chunkhash {
						fmt.Println("chunk data is lost .")
						continue
					} else {
						notvalidchnkdata = false

						_, err := file.Write(res.ChunkData)
						if err != nil {
							fmt.Println("error in write chunk to file : ", err)
						}
						chunkcount++
						break
					}
				} // end else
				if notvalidchnkdata { // currupted
					countnotvalidchnkdata++
					fmt.Println("Count of not valid chunk data :  ", countnotvalidchnkdata)
				}

			}
		}
	}
	fmt.Println("written chunk ", chunkcount)
	file0, er2 := ioutil.ReadFile(newPath)
	if er2 != nil {
		fmt.Println("error in  reading file !!!")
	}
	collectedhashfile := globalPkg.GetHash(file0)
	fmt.Println("Collected File Hash ", collectedhashfile)
	fmt.Println("Original File Hash  ", fStrct.FileHash2)

	// if collectedhashfile != fStrct.FileHash {
	// 	if countnotvalidchnkdata > 0 {
	// 		fmt.Println("error in getting chunk data !!!")
	// 	}
	// 	globalPkg.SendError(w, "server is down !")
	// 	globalPkg.WriteLog(logobj, "collected file hash not equall", "failed")
	// 	return
	// }

	// read file as bytes
	file2, er2 := os.Open(newPath)
	if er2 != nil {
		fmt.Println("error in  reading file !!!")
	}

	fileinfoCollected, _ := file2.Stat()
	fmt.Println("File Size           ", fStrct.FileSize)
	fmt.Println("Collected File Size ", fileinfoCollected.Size())
	// if fStrct.FileSize != fileinfoCollected.Size() {
	// 	globalPkg.SendError(w, "file is corrupted")
	// 	globalPkg.WriteLog(logobj, "file is corrupted size file is different", "failed")
	// 	return
	// }
	// ip := strings.Split(validator.CurrentValidator.ValidatorIP, ":")
	// fmt.Println("length of string :  ", len(ip))
	// strip := ip[0] + "s"
	// httpsip := strip + ":" + ip[1] + ":" + ip[2]
	// // u, err := url.Parse(validator.CurrentValidator.ValidatorIP)
	// u, err := url.Parse(httpsip)
	// fmt.Println("=================== link ", u, "========path ====  ", validator.CurrentValidator.ValidatorIP)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// u, err := url.Parse("https://us-demoinochain.inovatian.com")
	u, err := url.Parse(globalPkg.GlobalObj.Downloadfileip)

	u.Path = path.Join(u.Path, "files", fStrct.Fileid+fStrct.FileType)
	link := u.String()
	globalPkg.SendResponseMessage(w, link)
	globalPkg.WriteLog(logobj, "File downloaded successfully", "failed")

}
