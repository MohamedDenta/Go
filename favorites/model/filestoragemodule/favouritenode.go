package filestoragemodule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"../account"
	"../accountdb"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../globalPkg"
	"../logpkg"
	"../validator"
)

func verifyaccount(publickey *string, w *http.ResponseWriter, logobj *logpkg.LogStruct) (acc accountdb.AccountStruct) {
	acc = account.GetAccountByAccountPubicKey(*publickey)
	useraddress := acc.AccountPublicKey
	if useraddress != *publickey {
		globalPkg.SendError(*w, "error in public key")
		globalPkg.WriteLog(*logobj, "error in public key", "failed")
	}
	return acc
}
func verifysignature(useraddress *string, data *FavouriteNodesData, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
	validSig := false
	publickey := account.FindpkByAddress(*useraddress).Publickey
	if publickey != "" {
		parsedpublickey := cryptogrpghy.ParsePEMtoRSApublicKey(publickey)
		nodeids := strings.Join(data.NodeIDs, "")
		fmt.Println("nodeids :  ", nodeids)
		signatureData := data.PublicKey + nodeids // public key + nodeids
		//fmt.Println("Sig:- ", signatureData)
		validSig = cryptogrpghy.VerifyPKCS1v15(data.Signture, signatureData, *parsedpublickey)
	} else {
		validSig = false
	}
	validSig = true
	if !validSig {
		globalPkg.SendError(*w, "you are not allowed to do this operation")
		globalPkg.WriteLog(*logobj, "you are not allowed to do this operation", "failed")
	}
	return validSig
}
func verifynodes(data *FavouriteNodesData, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {

	//check num of nodes
	idscount := len(data.NodeIDs)
	if idscount == 0 {
		return true
	}
	if idscount > 5 {
		globalPkg.SendError(*w, "nodes count should be between 1 and 5 nodes")
		globalPkg.WriteLog(*logobj, "nodes count should be between 1 and 5 nodes", "failed")
		return false
	}
	nodes := validator.GetAllValidatorsDecrypted()
	found := false
	//check ids
	for _, id := range data.NodeIDs {
		found = false
		for _, node := range nodes {
			if node.ValidatorIP == id {
				found = true
				break
			}
		}
		if !found {
			globalPkg.SendError(*w, "incorrect node ids")
			globalPkg.WriteLog(*logobj, "incorrect node ids", "failed")
			return false
		}
	}

	return true
}
func addfavorednodes(accountObj *accountdb.AccountStruct, data *FavouriteNodesData) {
	accountObj.Storage = data.Storage
	for _, v := range data.NodeIDs {
		if len(accountObj.FavoredNodes) == 0 {
			accountObj.FavoredNodes = make(map[string]string)
		}
		accountObj.FavoredNodes[v] = v
	}
	broadcastTcp.BoardcastingTCP(*accountObj, "updateaccountFavoredNodes", "file")
}
func verifystorage(data *FavouriteNodesData, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
	// POSITIVE STORAGE
	if data.Storage < 1 {
		globalPkg.SendError(*w, "please enter a storage more than zero")
		globalPkg.WriteLog(*logobj, "please enter your correct request", "failed")
		return false
	}

	//TODO: NODES SATISFY THE STORAGE

	return true
}
func removefavorednodes(accountObj *accountdb.AccountStruct, data *FavouriteNodesData, w *http.ResponseWriter, logobj *logpkg.LogStruct) bool {
	var nodes map[string]string
	data.Storage = accountObj.Storage
	//n := len(accountObj.FavoredNodes)
	m := len(data.NodeIDs)
	// var found bool
	for i := 0; i < m; i++ {
		id, ok := accountObj.FavoredNodes[data.NodeIDs[i]]
		if id == "" || !ok {
			globalPkg.SendError(*w, "these nodes are not in your favourit list")
			globalPkg.WriteLog(*logobj, "these nodes are not in your favourit list", "failed")
			return false
		}
		delete(accountObj.FavoredNodes, data.NodeIDs[i])
	}
	// TODO : check nodes' space satisfy

	accountObj.FavoredNodes = nodes
	broadcastTcp.BoardcastingTCP(*accountObj, "updateaccountFavoredNodes", "file")
	return true
}

//AddFavouriteNode get sharefile from database
func AddFavouriteNode(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "AddFavouriteNode", "file", "_", "_", "_", 0}
	data := new(FavouriteNodesData)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(data)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	//VERIFY_STORAGE
	fine := verifystorage(data, &w, &logobj)
	if !fine {
		return
	}
	// VERIFY_ACCOUNT
	accountObj := verifyaccount(&(data.PublicKey), &w, &logobj)
	if accountObj.AccountPublicKey != data.PublicKey {
		return
	}

	// VERIFY_NODEIDs
	if !verifynodes(data, &w, &logobj) {
		return
	}

	// VERIFY_SIGNATURE
	if !verifysignature(&accountObj.AccountPublicKey, data, &w, &logobj) {
		return
	}

	// ADD_FAVORED_LIST_TO_USER
	addfavorednodes(&accountObj, data)

	globalPkg.SendResponseMessage(w, "you set favored nodes successfully")
	globalPkg.WriteLog(logobj, "you set favored nodes successfully", "success")
}

// RemoveFavouriteNode get sharefile from database
func RemoveFavouriteNode(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "RemoveFavouriteNode", "file", "_", "_", "_", 0}
	data := new(FavouriteNodesData)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(data)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	// VERIFY_ACCOUNT
	accountObj := verifyaccount(&(data.PublicKey), &w, &logobj)
	if accountObj.AccountPublicKey != data.PublicKey {
		return
	}

	// VERIFY_SIGNATURE
	if !verifysignature(&accountObj.AccountPublicKey, data, &w, &logobj) {
		return
	}

	// REMOVE_FAVORED_LIST_FROM_USER
	fine := removefavorednodes(&accountObj, data, &w, &logobj)
	if !fine {
		return
	}
	globalPkg.SendResponseMessage(w, "you removed favored nodes successfully")
	globalPkg.WriteLog(logobj, "you removed favored nodes successfully", "success")
}

// ExploreFavouriteNodes get sharefile from database
func ExploreFavouriteNodes(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ExploreFavouriteNodes", "file", "_", "_", "_", 0}
	data := new(FavouriteNodesData)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(data)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	accountObj := verifyaccount(&(data.PublicKey), &w, &logobj)
	if accountObj.AccountPublicKey != data.PublicKey {
		return
	}
	var res ExploreResp
	for key, _ := range accountObj.FavoredNodes {
		fmt.Println(key)
		res.Nodes = append(res.Nodes, key)
	}
	b, err := json.Marshal(res)
	if err != nil {
		globalPkg.SendError(w, "error in reading data")
		globalPkg.WriteLog(logobj, "error in reading data", "failed")
		return
	}
	globalPkg.SendResponse(w, b)
	globalPkg.WriteLog(logobj, "you explored favored nodes successfully", "success")

}
