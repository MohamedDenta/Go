package filestoragemodule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"../accountdb"
	"../filestorage"

	"../account"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../globalPkg"
	"../logpkg"
)

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
	validSig = true
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
	validSig = true
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
