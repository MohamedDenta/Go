package filestoragemodule

import (
	"encoding/json"
	"net/http"

	"../admin"
	"../filestorage"
	"../globalPkg"
)

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
