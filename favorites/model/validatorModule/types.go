package validatorModule

import (
	"encoding/json"
	"net/http"

	"../admin"
	"../validator"
)

//MixedObjec of admin object and validator
type MixedObjec struct {
	Admn  admin.Admin
	Vldtr validator.ValidatorStruct
}

// Decode decode MixedObjec from json
func (mo *MixedObjec) Decode(req *http.Request) error {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(mo)
	return err
}
