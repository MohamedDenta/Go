package filestoragemodule

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"../block"
	"../transaction"

	"../admin"

	"../filestorage"

	"time"

	"fmt"
	"math"
	"path/filepath"
	"strconv"

	"../account"
	"../accountdb"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../globalPkg"
	globalpkg "../globalPkg"
	"../logpkg"
	"../validator"
)

//Fileapi data from front , for front only
type Fileapi struct {
	Fileid   string
	Ownerpk  string
	FileName string
	FileType string
	FileHash string
	Timefile time.Time
	Signture string
}

const maxUploadSize = 5 * 1024 * 1024 * 1024 // 5 GB
const uploadPath = "upload"

// ExploreResponse explore response body
type ExploreResponse struct {
	OwnedFiles      []accountdb.FileList
	TotalSizeOwned  int64
	SharedFile      []accountdb.FileList
	TotalSizeShared int64
}

// ExploreBody expore request body
type ExploreBody struct {
	Publickey string
	Password  string
}

// RetrieveBody retrieve request body
type RetrieveBody struct {
	Publickey string
	Password  string
	FileID    string
	Time      string
	Signture  string
}

// ShareFiledata share file
type ShareFiledata struct {
	Publickey        string
	Password         string
	FileID           string
	PermissionPkList []string
	Signture         string
}

//UploadFile upload file
func UploadFile(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UploadFile", "file", "_", "_", "_", 0}
	filestructObj := filestorage.FileStruct{}
	// validate file size
	// r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	// if err := r.ParseMultipartForm(maxUploadSize); err != nil {
	// 	globalPkg.SendError(w, "file too big")
	// 	globalPkg.WriteLog(logobj, "file too big", "failed")
	// 	return
	// }

	// contentType := r.Header.Get("Content-Type")
	// fmt.Println("Content Type ", contentType)

	// parse and validate file and post parameters
	file, fileInfo, err := r.FormFile("uploadFile")
	if err != nil {
		globalPkg.SendError(w, "invalid file")
		globalPkg.WriteLog(logobj, "invalid file", "failed")
		return
	}
	defer file.Close()
	// validate file size
	if fileInfo.Size > maxUploadSize {
		globalPkg.SendError(w, "file too big")
		globalPkg.WriteLog(logobj, "file too big", "failed")
		return
	}
	// read file as bytes
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		globalPkg.SendError(w, "invalid file")
		globalPkg.WriteLog(logobj, "invalid file", "failed")
		return
	}

	// filename with extension
	filestructObj.FileName = r.FormValue("FileName")
	if filestructObj.FileName != fileInfo.Filename {
		globalPkg.SendError(w, "file name does not match + extension")
		globalPkg.WriteLog(logobj, "file name does not match", "failed")
		return
	}
	//Extension of file
	fileextenstion := filepath.Ext(filestructObj.FileName) //extension with .

	filestructObj.FileType = r.FormValue("FileType")
	if filestructObj.FileType != fileextenstion {
		globalPkg.SendError(w, "file extension does not match")
		globalPkg.WriteLog(logobj, "file extension does not match", "failed")
		return
	}

	//check of hash file from front and get hash 256
	hashfile := globalPkg.GetHash(fileBytes)
	fmt.Println("hash file :   ", hashfile)
	filestructObj.FileHash = r.FormValue("FileHash")
	filestructObj.FileHash2 = hashfile // for security issue
	// if filestructObj.FileHash != hashfile {
	// 	globalPkg.SendError(w, "file hash does not  match")
	// 	globalPkg.WriteLog(logobj, "file hash does not  match", "failed")
	// 	return
	// }

	//check for time
	timef := r.FormValue("Timefile")

	filestructObj.Timefile, _ = time.Parse("2006-01-02T15:04:05Z07:00", timef) //convert string to time

	// time differnce between the received file time and the server's time.
	// tnow := globalPkg.UTCtime()
	// tfile := globalPkg.UTCtimefield(filestructObj.Timefile)
	// timeDifference := tnow.Sub(tfile).Seconds()
	// // fmt.Println("  Time Difference    :     ", timeDifference ,"-------now   -   ", tnow , "    tfile  ****     ", tfile )
	// if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
	// 	globalPkg.SendError(w, "please check your time")
	// 	globalPkg.WriteLog(logobj, "please check your time", "failed")
	// 	return
	// }

	//check for pk is exist in account
	filestructObj.Ownerpk = r.FormValue("Ownerpk")
	accountObj := account.GetAccountByAccountPubicKey(filestructObj.Ownerpk)
	if accountObj.AccountPublicKey == "" {
		globalPkg.SendError(w, "public key  not exist")
		globalPkg.WriteLog(logobj, "pk not exist", "failed")
		return
	}
	var totalsize int64 = fileInfo.Size
	//check for this user all file list is less than 5GB
	for _, file := range accountObj.Filelist {
		totalsize += file.FileSize
	}
	if totalsize > maxUploadSize {
		globalPkg.SendError(w, "sorry your uploaded storage exceeded 5 GB ")
		globalPkg.WriteLog(logobj, "sorry your uploaded storage exceeded 5 GB ", "failed")
		return
	}
	// Signture string
	pk := account.FindpkByAddress(accountObj.AccountPublicKey).Publickey
	validSig := false
	if pk != "" {
		publickey := cryptogrpghy.ParsePEMtoRSApublicKey(pk)

		// signatureData := filestructObj.FileName + filestructObj.FileType +
		// 	filestructObj.FileHash + filestructObj.Ownerpk + timef
		signatureData := filestructObj.FileName + filestructObj.FileType +
			filestructObj.Ownerpk
		signature := r.FormValue("Signture")
		validSig = cryptogrpghy.VerifyPKCS1v15(signature, signatureData, *publickey)
		// validSig = true
	} else {
		validSig = false
	}
	if validSig {
		fmt.Println("")
		//return ""
		// } else if !validSig {
		// 	fmt.Println("")
	} else {
		globalPkg.SendError(w, "You are not allowed to upload file")
		globalPkg.WriteLog(logobj, "You are not allowed to upload file", "failed")
		return
	}
	//generate file id
	filestructObj.Fileid = FileIndex(accountObj)

	// validator active and not remove
	var validatorlistactive []validator.ValidatorStruct
	validatorList := validator.GetAllValidators()
	for _, validatorObj := range validatorList {
		if validatorObj.ValidatorActive == true && validatorObj.ValidatorRemove == false {
			validatorlistactive = append(validatorlistactive, validatorObj)
		}
	}
	var resp broadcastTcp.FileBroadcastResponse
	var fileSize int64 = fileInfo.Size
	filestructObj.FileSize = fileSize
	const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
	fileReader, errf := fileInfo.Open()
	if errf != nil {
		fmt.Println("error in reading file ", errf)
	}
	for i := uint64(0); i < totalPartsNum; i++ {

		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		fileReader.Read(partBuffer)

		chunkObj := filestorage.Chunkdb{}
		// change chunkid , file id to hash and auto increment
		chunkObj.Chunkid = filestructObj.Fileid + "_" + strconv.Itoa(int(i))
		chunkObj.Fileid = filestructObj.Fileid
		chunkObj.Chunkhash = globalPkg.GetHash(partBuffer)
		chunkObj.ChunkNumber = int(i)
		chunkObj.Chunkdata = partBuffer

		chunkdataObj := filestorage.Chunkdata{}
		var chunks []filestorage.Chunkdata
		chunkdataObj.Chunkhash = chunkObj.Chunkhash
		validatorindex := randomValidator(len(validatorlistactive))
		fmt.Println("***********************validator list actives ", validatorindex)
		for _, i := range validatorindex {
			// time.Sleep(time.Millisecond * 10)
			validObj := validatorlistactive[i]
			if validObj.ValidatorPublicKey == validator.CurrentValidator.ValidatorPublicKey {
				_, _, resp = broadcastTcp.SendObject(chunkObj, validator.CurrentValidator.ValidatorPublicKey, "addchunk", "file", validator.CurrentValidator.ValidatorSoketIP)
			} else {
				_, _, resp = broadcastTcp.SendObject(chunkObj, validObj.ValidatorPublicKey, "addchunk", "file", validObj.ValidatorSoketIP)
			}
			time.Sleep(time.Millisecond * 10)
			fmt.Println("    *********************************chan ", resp.Valid)
			if resp.Valid {

				chunkdataObj.ValidatorIP = validObj.ValidatorIP
				chunks = append(chunks, chunkdataObj)
			}
		}
		if filestructObj.Mapping == nil {
			filestructObj.Mapping = make(map[string][]filestorage.Chunkdata)
		}
		if chunks == nil {
			globalPkg.SendError(w, "failed to upload file. try to upload it again")
			globalPkg.WriteLog(logobj, "failed to upload file", "failed")
			return
		}
		filestructObj.Mapping[chunkObj.Chunkid] = chunks
	}

	//create transaction id
	filestructObj.Transactionid = globalPkg.CreateHash(filestructObj.Timefile, fmt.Sprintf("%s", filestructObj), 3)
	// add on trsansaction pool
	broadcastTcp.BoardcastingTCP(filestructObj, "addfile", "file")
	globalPkg.SendResponseMessage(w, "File Upload Successfully")
	globalPkg.WriteLog(logobj, "get balance success", "success")

}

// ExploreFiles explore all file for some user
func ExploreFiles(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ExploreFiles", "file", "_", "_", "_", 0}
	var obj ExploreBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&obj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	time.Sleep(time.Millisecond * 10) // for handle unknown issue
	acc := account.GetAccountByAccountPubicKey(obj.Publickey)
	if acc.AccountPublicKey != obj.Publickey {
		globalPkg.SendError(w, "error in public key")
		globalPkg.WriteLog(logobj, "error in public key", "failed")
		return
	}
	if acc.AccountPassword != obj.Password {
		globalPkg.SendError(w, "error in password")
		globalPkg.WriteLog(logobj, "error in password", "failed")
		return
	}
	FilesData := ExploreResponse{}
	files := acc.Filelist
	for _, file := range files {
		FilesData.TotalSizeOwned += file.FileSize
		FilesData.OwnedFiles = append(FilesData.OwnedFiles, file)
	}

	sharefiles := filestorage.FindSharedfileByAccountIndex(acc.AccountIndex)
	// if sharefiles.AccountIndex == "" {
	// 	fmt.Println("-----------------not take share file ----------")
	// }
	if len(sharefiles.OwnerSharefile) != 0 {
		for _, sharefileObj := range sharefiles.OwnerSharefile {
			accountObj := account.GetAccountByAccountPubicKey(sharefileObj.OwnerPublicKey)
			for _, filelistObj := range accountObj.Filelist {
				if containsfileid(sharefileObj.Fileid, filelistObj.Fileid) {
					FilesData.TotalSizeShared += filelistObj.FileSize
					FilesData.SharedFile = append(FilesData.SharedFile, filelistObj)
				}
			}
		}
	}
	//check files in transaction pool with status deleted
	txs := transaction.Pending_transaction
	for _, tx := range txs {
		// owned files => owner delete files and get explore files
		if tx.SenderPK == acc.AccountPublicKey {
			fileObj := tx.Transaction.Filestruct
			if fileObj.FileSize != 0 && fileObj.Deleted == true {
				indexfileid := containsfileidinfilelist(FilesData.OwnedFiles, fileObj.Fileid)
				if indexfileid != -1 {
					FilesData.TotalSizeOwned -= fileObj.FileSize
					FilesData.OwnedFiles = append(FilesData.OwnedFiles[:indexfileid], FilesData.OwnedFiles[indexfileid+1:]...)
				}
			}
		}
		//check for shared files
		fileObj1 := tx.Transaction.Filestruct
		if fileObj1.FileSize != 0 && fileObj1.Deleted == true {
			indexfileidshare := containsfileidinfilelist(FilesData.SharedFile, fileObj1.Fileid)
			if indexfileidshare != -1 {
				FilesData.TotalSizeShared -= fileObj1.FileSize
				FilesData.SharedFile = append(FilesData.SharedFile[:indexfileidshare], FilesData.SharedFile[indexfileidshare+1:]...)
			}
		}
	}

	sendJSON, _ := json.Marshal(FilesData)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get files success", "success")
}

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
		// validSig = true
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
	decryptIndexBlock1 := cryptogrpghy.KeyDecrypt(globalpkg.EncryptAccount, selectedFile.Blockindex)
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
	validatorLst := validator.GetAllValidators()
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
				if validatorObj2.ValidatorPublicKey == validator.CurrentValidator.ValidatorPublicKey {
					_, _, res = broadcastTcp.SendObject(chnkObj, validator.CurrentValidator.ValidatorPublicKey, "getchunkdata", "file", validator.CurrentValidator.ValidatorSoketIP)

				} else {
					_, _, res = broadcastTcp.SendObject(chnkObj, validatorObj2.ValidatorPublicKey, "getchunkdata", "file", validatorObj2.ValidatorSoketIP)
				}
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

// DeleteFile delete a specific file
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "DeleteFile", "file", "_", "_", "_", 0}
	var obj RetrieveBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&obj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	time.Sleep(time.Millisecond * 10)
	acc := account.GetAccountByAccountPubicKey(obj.Publickey)
	if acc.AccountPublicKey != obj.Publickey {
		globalPkg.SendError(w, "error in public key")
		globalPkg.WriteLog(logobj, "error in public key", "failed")
		return
	}
	if acc.AccountPassword != obj.Password {
		globalPkg.SendError(w, "error in password")
		globalPkg.WriteLog(logobj, "error in password", "failed")
		return
	}
	// 	tnow := globalPkg.UTCtime()
	// 	t, _ := time.Parse("2006-01-02T15:04:05Z07:00", obj.Time)
	// 	tfile := globalPkg.UTCtimefield(t)
	// 	timeDifference := tnow.Sub(tfile).Seconds()
	// 	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
	// 		globalPkg.SendError(w, "please check your time")
	// 		globalPkg.WriteLog(logobj, "please check your time", "failed")
	// 		return
	// 	}

	// Signture string
	validSig := false
	pk := account.FindpkByAddress(acc.AccountPublicKey).Publickey
	if pk != "" {
		publickey := cryptogrpghy.ParsePEMtoRSApublicKey(pk)
		// signatureData := obj.FileID + obj.Publickey + obj.Password + obj.Time
		signatureData := obj.Publickey + obj.Password + obj.FileID
		validSig = cryptogrpghy.VerifyPKCS1v15(obj.Signture, signatureData, *publickey)
		// validSig = true
	} else {
		validSig = false
	}
	if !validSig {
		globalPkg.SendError(w, "you are not allowed to delete")
		globalPkg.WriteLog(logobj, "you are not allowed to delete", "failed")
		return
	}
	// check user own this file id
	files := acc.Filelist
	found := false
	foundtx := false
	var selectedFile accountdb.FileList
	for _, file := range files {
		if file.Fileid == obj.FileID {
			found = true
			selectedFile = file
			break
		}
	}
	if !found {
		globalPkg.SendError(w, "You don't have this file")
		globalPkg.WriteLog(logobj, "You don't have this file", "failed")
		return
	}

	//check files in transaction pool with status deleted
	txs := transaction.Pending_transaction
	for _, tx := range txs {
		// owned files => owner delete files and get explore files
		if tx.SenderPK == acc.AccountPublicKey {
			fileObj := tx.Transaction.Filestruct
			if fileObj.FileSize != 0 && fileObj.Deleted == true && fileObj.Fileid == obj.FileID {
				foundtx = true
				break
			}
		}
	}
	if foundtx {
		globalPkg.SendError(w, "You don't have this file!")
		globalPkg.WriteLog(logobj, "You don't have this file", "failed")
		return
	}
	decryptIndexBlock1 := cryptogrpghy.KeyDecrypt(globalpkg.EncryptAccount, selectedFile.Blockindex)
	fmt.Println("Block Index to be deletd data ", decryptIndexBlock1)
	blkObj := block.GetBlockInfoByID(decryptIndexBlock1)
	var fStrct filestorage.FileStruct
	for _, tx := range blkObj.BlockTransactions {
		fStrct = tx.Filestruct
		if fStrct.Fileid == selectedFile.Fileid {
			fStrct = tx.Filestruct
			break
		}
	}
	fStrct.Deleted = true
	fStrct.Transactionid = globalPkg.CreateHash(fStrct.Timefile, fmt.Sprintf("%s", fStrct), 3)
	// add on trsansaction pool
	broadcastTcp.BoardcastingTCP(fStrct, "deletefile", "file")
	// delete file if share from share file table

	globalPkg.SendResponseMessage(w, "File Deleted Successfully")
	globalPkg.WriteLog(logobj, "File Deleted Successfully", "success")
}

//GetAllChunksAPI get chunks from database
func GetAllChunksAPI(w http.ResponseWriter, req *http.Request) {

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		jsonObj, _ := json.Marshal(filestorage.GetAllChunks())
		globalPkg.SendResponse(w, jsonObj)
	} else {
		globalPkg.SendError(w, "you are not the admin ")
	}
}

//ShareFiles share  file for some user request file id , pk, password , list of pk to be shared
func ShareFiles(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ShareFiles", "file", "_", "_", "_", 0}
	var ShareFiledataObj ShareFiledata
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&ShareFiledataObj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	time.Sleep(time.Millisecond * 10) // for handle unknown issue
	accountObj := account.GetAccountByAccountPubicKey(ShareFiledataObj.Publickey)
	if accountObj.AccountPublicKey != ShareFiledataObj.Publickey {
		globalPkg.SendError(w, "error in public key")
		globalPkg.WriteLog(logobj, "error in public key", "failed")
		return
	}
	if accountObj.AccountPassword != ShareFiledataObj.Password {
		globalPkg.SendError(w, "error in password")
		globalPkg.WriteLog(logobj, "error in password", "failed")
		return
	}
	// check user own this file id
	files := accountObj.Filelist
	found := false
	for _, fileObj := range files {
		if fileObj.Fileid == ShareFiledataObj.FileID {
			found = true
		}
	}
	if !found {
		globalPkg.SendError(w, "You don't have this file")
		globalPkg.WriteLog(logobj, "You don't have this file", "failed")
		return
	}
	// check pk already exist in blockchain
	accountList := accountdb.GetAllAccounts()
	for _, pk := range ShareFiledataObj.PermissionPkList {
		if !containspk(accountList, pk) {
			globalPkg.SendError(w, "this public key is not associated with any account")
			globalPkg.WriteLog(logobj, "You don't have this file", "failed")
			return
		}
	}
	// Signture string
	validSig := false
	pk1 := account.FindpkByAddress(accountObj.AccountPublicKey).Publickey
	if pk1 != "" {
		publickey1 := cryptogrpghy.ParsePEMtoRSApublicKey(pk1)
		strpermissionlist := strings.Join(ShareFiledataObj.PermissionPkList, "")
		fmt.Println("strpermissionlist :  ", strpermissionlist)
		signatureData := strpermissionlist + ShareFiledataObj.FileID + ShareFiledataObj.Publickey
		validSig = cryptogrpghy.VerifyPKCS1v15(ShareFiledataObj.Signture, signatureData, *publickey1)
	} else {
		validSig = false
	}
	// validSig = true
	if !validSig {
		globalPkg.SendError(w, "you are not allowed to share file")
		globalPkg.WriteLog(logobj, "you are not allowed to share file", "failed")
		return
	}
	//

	filelistOwner := accountObj.Filelist
	// add account index see file , ownerpk , fileid
	//append share file id , ownerpk to account index want to share file to you
	for _, pk := range ShareFiledataObj.PermissionPkList {
		var sharedfileObj filestorage.SharedFile
		var ownerfileObj filestorage.OwnersharedFile
		var ownerfileObj2 filestorage.OwnersharedFile
		var foundOwnerpk bool
		accountind := account.GetAccountByAccountPubicKey(pk)
		sharedfileObj.AccountIndex = accountind.AccountIndex
		ownedsharefile := filestorage.FindSharedfileByAccountIndex(sharedfileObj.AccountIndex)
		if pk != ShareFiledataObj.Publickey { //same owner share to himself
			if len(ownedsharefile.OwnerSharefile) != 0 {
				for _, ownedsharefileObj := range ownedsharefile.OwnerSharefile {

					if ownedsharefileObj.OwnerPublicKey == ShareFiledataObj.Publickey {
						foundOwnerpk = true
						if !containsfileid(ownedsharefileObj.Fileid, ShareFiledataObj.FileID) {
							ownedsharefileObj.Fileid = append(ownedsharefileObj.Fileid, ShareFiledataObj.FileID)
						}
					}
					sharedfileObj.OwnerSharefile = append(sharedfileObj.OwnerSharefile, ownedsharefileObj)
				}
				if !foundOwnerpk {
					ownerfileObj2.OwnerPublicKey = ShareFiledataObj.Publickey
					ownerfileObj2.Fileid = append(ownerfileObj2.Fileid, ShareFiledataObj.FileID)
					sharedfileObj.OwnerSharefile = append(sharedfileObj.OwnerSharefile, ownerfileObj2)
				}

			} else {
				ownerfileObj.OwnerPublicKey = ShareFiledataObj.Publickey
				ownerfileObj.Fileid = append(ownerfileObj.Fileid, ShareFiledataObj.FileID)
				sharedfileObj.OwnerSharefile = append(sharedfileObj.OwnerSharefile, ownerfileObj)
			}
			broadcastTcp.BoardcastingTCP(sharedfileObj, "sharefile", "file")
			//append permisssionlist to account owner filelist
			for m := range filelistOwner {
				if filelistOwner[m].Fileid == ShareFiledataObj.FileID {
					if !containsfileid(filelistOwner[m].PermissionList, pk) {
						filelistOwner[m].PermissionList = append(filelistOwner[m].PermissionList, pk)
					}
					break
				}
			}
		}
	}
	accountObj.Filelist = filelistOwner
	broadcastTcp.BoardcastingTCP(accountObj, "updateaccountFilelist", "file")

	globalPkg.SendResponseMessage(w, "you shared file successfully")
	globalPkg.WriteLog(logobj, "you shared file successfully", "success")
}

//UnshareFile unshare a specific file delete it from share files table
func UnshareFile(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UnshareFile", "file", "_", "_", "_", 0}
	var requestObj RetrieveBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&requestObj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	time.Sleep(time.Millisecond * 10)
	accountObj := account.GetAccountByAccountPubicKey(requestObj.Publickey)
	if accountObj.AccountPublicKey != requestObj.Publickey {
		globalPkg.SendError(w, "error in public key")
		globalPkg.WriteLog(logobj, "error in public key", "failed")
		return
	}
	if accountObj.AccountPassword != requestObj.Password {
		globalPkg.SendError(w, "error in password")
		globalPkg.WriteLog(logobj, "error in password", "failed")
		return
	}
	// Signture string
	validSig := false
	pk1 := account.FindpkByAddress(accountObj.AccountPublicKey).Publickey
	if pk1 != "" {
		publickey1 := cryptogrpghy.ParsePEMtoRSApublicKey(pk1)
		signatureData := requestObj.Publickey + requestObj.Password + requestObj.FileID
		validSig = cryptogrpghy.VerifyPKCS1v15(requestObj.Signture, signatureData, *publickey1)
	} else {
		validSig = false
	}
	// validSig = true
	if !validSig {
		globalPkg.SendError(w, "you are not allowed to delete unshare file")
		globalPkg.WriteLog(logobj, "you are not allowed to delete unshare file", "failed")
		return
	}
	found := false
	sharefile := filestorage.FindSharedfileByAccountIndex(accountObj.AccountIndex)
	if len(sharefile.OwnerSharefile) != 0 {
		for sharefileindex, sharefileObj := range sharefile.OwnerSharefile {
			fileindex := containsfileidindex(sharefileObj.Fileid, requestObj.FileID)
			if fileindex != -1 {
				found = true
				sharefileObj.Fileid = append(sharefileObj.Fileid[:fileindex], sharefileObj.Fileid[fileindex+1:]...)
				sharefile.OwnerSharefile = append(sharefile.OwnerSharefile[:sharefileindex], sharefile.OwnerSharefile[sharefileindex+1:]...)
				// fmt.Println("============== file ids :", len(sharefileObj.Fileid), "============", len(sharefile.OwnerSharefile))
				// delete from permission list
				accountOwnerObj := account.GetAccountByAccountPubicKey(sharefileObj.OwnerPublicKey)
				FilelistOwner := accountOwnerObj.Filelist
				var indexpk int = -1
				var indexfile int = -1
				for j, fileOwnerObj := range FilelistOwner {
					if fileOwnerObj.Fileid == requestObj.FileID {
						if len(fileOwnerObj.PermissionList) != 0 {
							for k, pkpermission := range fileOwnerObj.PermissionList {
								if pkpermission == requestObj.Publickey {
									indexpk = k
									indexfile = j
									break
								}
							}
						}
					}
				}

				if indexpk != -1 {
					accountOwnerObj.Filelist[indexfile].PermissionList = append(accountOwnerObj.Filelist[indexfile].PermissionList[:indexpk], accountOwnerObj.Filelist[indexfile].PermissionList[indexpk+1:]...)
					broadcastTcp.BoardcastingTCP(accountOwnerObj, "updateaccountFilelist", "file")
					// accountOwnerObj.Filelist = FilelistOwner
				}
				//
				if len(sharefileObj.Fileid) != 0 && len(sharefile.OwnerSharefile) >= 1 {
					sharefile.OwnerSharefile = append(sharefile.OwnerSharefile, sharefileObj)
				} else if len(sharefileObj.Fileid) != 0 && len(sharefile.OwnerSharefile) == 0 {
					sharefile.OwnerSharefile = append(sharefile.OwnerSharefile, sharefileObj)
				}
				broadcastTcp.BoardcastingTCP(sharefile, "updatesharefile", "file")

				if len(sharefile.OwnerSharefile) == 0 {
					broadcastTcp.BoardcastingTCP(sharefile, "deleteaccountindex", "file")
				}
				globalPkg.SendResponseMessage(w, "you unshare file successfully")
				globalPkg.WriteLog(logobj, "you unshare file successfully", "success")
				return

			}
		}
	}

	if !found {
		globalPkg.SendError(w, "you not take share file")
		globalPkg.WriteLog(logobj, "you not take share file", "failed")
		return
	}

}

//GetAllShareFileAPI get sharefile from database
func GetAllShareFileAPI(w http.ResponseWriter, req *http.Request) {

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		jsonObj, _ := json.Marshal(filestorage.GetAllSharedFile())
		globalPkg.SendResponse(w, jsonObj)
	} else {
		globalPkg.SendError(w, "you are not the admin ")
	}
}
