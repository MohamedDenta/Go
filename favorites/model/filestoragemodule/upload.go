package filestoragemodule

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"../accountdb"
	"../filestorage"

	"time"

	"fmt"
	"math"
	"path/filepath"
	"strconv"

	"../account"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../globalPkg"
	"../logpkg"
	"../validator"
)

func validatefilesize(r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
	// r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	// if err := r.ParseMultipartForm(maxUploadSize); err != nil {
	// 	globalPkg.SendError(w, "file too big")
	// 	globalPkg.WriteLog(logobj, "file too big", "failed")
	// 	return false
	// }
	return true
}
func readfile(r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) (bool, []byte, *multipart.FileHeader) {
	var fileBytes []byte
	// parse and validate file and post parameters
	file, fileInfo, err := r.FormFile("uploadFile")
	if err != nil {
		globalPkg.SendError(*w, "invalid file")
		globalPkg.WriteLog(*logobj, "invalid file", "failed")
		return false, fileBytes, fileInfo
	}
	defer file.Close()
	// validate file size
	if fileInfo.Size > maxUploadSize {
		globalPkg.SendError(*w, "file too big")
		globalPkg.WriteLog(*logobj, "file too big", "failed")
		return false, fileBytes, fileInfo
	}

	// read file as bytes
	fileBytes, err = ioutil.ReadAll(file)
	if err != nil {
		globalPkg.SendError(*w, "invalid file")
		globalPkg.WriteLog(*logobj, "invalid file", "failed")
		return false, fileBytes, fileInfo
	}
	return true, fileBytes, fileInfo
}
func readfileinfo(fname string, r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) (bool, string, string) {
	var name, extension string
	// filename with extension
	name = r.FormValue("FileName")
	if name != fname {
		globalPkg.SendError(*w, "file name does not match + extension")
		globalPkg.WriteLog(*logobj, "file name does not match", "failed")
		return false, "", ""
	}
	//Extension of file
	fileextenstion := filepath.Ext(fname) //extension with .

	extension = r.FormValue("FileType")
	if extension != fileextenstion {
		globalPkg.SendError(*w, "file extension does not match")
		globalPkg.WriteLog(*logobj, "file extension does not match", "failed")
		return false, "", ""
	}

	return true, name, extension
}
func readfilehash(fhash string, r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) (bool, string) {

	fmt.Println("hash file :   ", fhash)
	hash := r.FormValue("FileHash")
	// if filestructObj.FileHash != hashfile {
	// 	globalPkg.SendError(*w, "file hash does not  match")
	// 	globalPkg.WriteLog(*logobj, "file hash does not  match", "failed")
	// 	return false ,hash
	// }
	return true, hash
}
func readtime(r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) (bool, time.Time) {
	timef := r.FormValue("Timefile")

	T, _ := time.Parse("2006-01-02T15:04:05Z07:00", timef) //convert string to time
	// if err != nil {
	// 	globalPkg.SendError(*w, "please check your time")
	// 	globalPkg.WriteLog(*logobj, "please check your time", "failed")
	// 	return false, T
	// }
	// time differnce between the received file time and the server's time.
	// tnow := globalPkg.UTCtime()
	// tfile := globalPkg.UTCtimefield(filestructObj.Timefile)
	// timeDifference := tnow.Sub(tfile).Seconds()
	// // fmt.Println("  Time Difference    :     ", timeDifference ,"-------now   -   ", tnow , "    tfile  ****     ", tfile )
	// if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
	// 	globalPkg.SendError(w, "please check your time")
	// 	globalPkg.WriteLog(logobj, "please check your time", "failed")
	// 	return false,T
	// }
	return true, T
}
func readpublickey(r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) (bool, accountdb.AccountStruct) {
	pk := r.FormValue("Ownerpk")
	accountObj := account.GetAccountByAccountPubicKey(pk)
	if accountObj.AccountPublicKey == "" {
		globalPkg.SendError(*w, "public key  not exist")
		globalPkg.WriteLog(*logobj, "pk not exist", "failed")
		return false, accountObj
	}
	return true, accountObj
}
func checktotalstorage(sz int64, accountObj *accountdb.AccountStruct, r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
	var totalsize int64 = sz
	//check for this user all file list is less than 5GB
	for _, file := range accountObj.Filelist {
		totalsize += file.FileSize
	}
	if totalsize > maxUploadSize {
		globalPkg.SendError(*w, "sorry your uploaded storage exceeded 5 GB ")
		globalPkg.WriteLog(*logobj, "sorry your uploaded storage exceeded 5 GB ", "failed")
		return false
	}
	return true
}
func validatesignature(accountObj *accountdb.AccountStruct, filestructObj *filestorage.FileStruct, r *http.Request, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
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
		validSig = true
	} else {
		validSig = false
	}
	if validSig {
		fmt.Println("")
		//return ""
		// } else if !validSig {
		// 	fmt.Println("")
	} else {
		globalPkg.SendError(*w, "You are not allowed to upload file")
		globalPkg.WriteLog(*logobj, "You are not allowed to upload file", "failed")
		return false
	}
	return true
}

func getactivenodes() []validator.ValidatorStruct {
	var validatorlistactive []validator.ValidatorStruct
	validatorList := validator.GetAllValidatorsDecrypted()
	for _, validatorObj := range validatorList {
		if validatorObj.ValidatorActive == true && validatorObj.ValidatorRemove == false {
			validatorlistactive = append(validatorlistactive, validatorObj)
		}
	}
	return validatorlistactive
}

//UploadFile upload file
func UploadFile(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UploadFile", "file", "_", "_", "_", 0}
	filestructObj := new(filestorage.FileStruct)

	// READ_FILE
	fine, filebytes, fileinfo := readfile(r, &w, &logobj)
	if !fine {
		return
	}
	fine, name, extension := readfileinfo(fileinfo.Filename, r, &w, &logobj)
	if !fine {
		return
	}
	// filename with extension
	filestructObj.FileName = name
	// file extension
	filestructObj.FileType = extension

	//check of hash file from front and get hash 256
	hashfile := globalPkg.GetHash(filebytes)
	fine, hash := readfilehash(hashfile, r, &w, &logobj)
	if !fine {
		return
	}

	filestructObj.FileHash = hash
	filestructObj.FileHash2 = hashfile // for security issue

	//check for time
	fine, filestructObj.Timefile = readtime(r, &w, &logobj)
	if !fine {
		return
	}

	//check for pk is exist in account
	fine, accountObj := readpublickey(r, &w, &logobj)
	filestructObj.Ownerpk = accountObj.AccountPublicKey
	if !fine {
		return
	}
	//check for this user all file list is less than 5GB
	fine = checktotalstorage(fileinfo.Size, &accountObj, r, &w, &logobj)
	if !fine {
		return
	}

	// Signture string
	fine = validatesignature(&accountObj, filestructObj, r, &w, &logobj)
	if !fine {
		return
	}

	//generate file id
	filestructObj.Fileid = FileIndex(accountObj)

	// validator active and not remove
	validatorlistactive := getactivenodes()
	if len(validatorlistactive) == 0 {
		fmt.Println("active nodes = 0")
	}

	// distibute file over prefered storage nodes
	fine = distributefile(filestructObj, fileinfo, &accountObj, validatorlistactive[:], &w, &logobj)
	if !fine {
		return
	}

	//create transaction id
	filestructObj.Transactionid = globalPkg.CreateHash(filestructObj.Timefile, fmt.Sprintf("%s", filestructObj), 3)
	// add on trsansaction pool
	broadcastTcp.BoardcastingTCP(filestructObj, "addfile", "file")
	globalPkg.SendResponseMessage(w, "File Upload Successfully")
	globalPkg.WriteLog(logobj, "get balance success", "success")

}

func distributefile(filestructObj *filestorage.FileStruct, fileinfo *multipart.FileHeader, accountObj *accountdb.AccountStruct, validatorlistactive []validator.ValidatorStruct, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
	var resp broadcastTcp.FileBroadcastResponse
	var fileSize int64 = fileinfo.Size
	filestructObj.FileSize = fileSize
	const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)
	fileReader, errf := fileinfo.Open()
	if errf != nil {
		fmt.Println("error in reading file ", errf)
	}
	// in case of no favored nodes
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
		var validatorindex []int
		// check if user has favored nodes
		if len(accountObj.FavoredNodes) > 0 {
			validatorindex = getfoavorednodes(validatorlistactive[:], accountObj.FavoredNodes)
		} else {
			validatorindex = randomValidator(len(validatorlistactive))
		}

		fmt.Println("***********************validator list actives ", validatorindex)
		for _, i := range validatorindex {
			validObj := validatorlistactive[i]
			_, _, resp = broadcastTcp.SendObject(chunkObj, validObj.ECCPublicKey, "addchunk", "file", validObj.ValidatorSoketIP)
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
			globalPkg.SendError(*w, "failed to upload file. try to upload it again")
			globalPkg.WriteLog(*logobj, "failed to upload file", "failed")
			return false
		}
		filestructObj.Mapping[chunkObj.Chunkid] = chunks
	}
	return true
}

func getfoavorednodes(validatorlistactive []validator.ValidatorStruct, favorednodes map[string]string) []int {
	var ret []int
	for i, v := range validatorlistactive {
		if favorednodes[v.ValidatorIP] == v.ValidatorIP {
			ret = append(ret, i)
		}

	}
	fmt.Println("ret ---------------- ", ret)
	return ret
}
