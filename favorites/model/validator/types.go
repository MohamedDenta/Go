package validator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"../admin"
	"../cryptogrpghy"
	"../globalPkg"
)

// AbstractNode interface for essential methods in any node
type AbstractNode interface {
}

//ValidatorStruct is the struct for db
type ValidatorStruct struct {
	ValidatorIP            string
	ValidatorSoketIP       string
	ECCPublicKey           ecdsa.PublicKey
	ECCPrivateKey          *ecdsa.PrivateKey
	EncECCPublic           string
	EncECCPriv             string
	ValidatorStakeCoins    float64
	ValidatorRegisterTime  time.Time
	ValidatorActive        bool
	ValidatorLastHeartBeat time.Time
	ValidatorRemove        bool
	Index                  string
}

// Decode read validator from json
func (v *ValidatorStruct) Decode(data []byte) error {
	err := json.Unmarshal(data, v)
	return err

}

//DecodeBody decode from request body
func (v *ValidatorStruct) DecodeBody(req *http.Request) error {
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	return err

}

//DecryptPK decrypt validator public key
func (v *ValidatorStruct) DecryptPK() string {
	return cryptogrpghy.KeyDecrypt(globalPkg.AESEncryptionKey, v.EncECCPublic)
}

//DecryptPrvK decrypt validator private key
func (v *ValidatorStruct) DecryptPrvK() string {
	return cryptogrpghy.KeyDecrypt(globalPkg.AESEncryptionKey, v.EncECCPriv)
}

//ECCPublicKeyStruct used in unmarshal objects that contains ECC public key
type ECCPublicKeyStruct struct {
	Key RetrieveECCPublicKey `json:"ECCPublicKey"`
}

// Decode read ecc public key from json
func (e *ECCPublicKeyStruct) Decode(data []byte) {
	json.Unmarshal(data, e)

}

//RetrieveECCPublicKey used to unmarshal ECC public key
type RetrieveECCPublicKey struct {
	CurveParams *elliptic.CurveParams `json:"Curve"`
	MyX         *big.Int              `json:"X"`
	MyY         *big.Int              `json:"Y"`
}

// Decode read ecc public key from json
func (e *RetrieveECCPublicKey) Decode(data []byte) {
	json.Unmarshal(data, e)

}

//ECCPrivateStruct used in unmarshal objects that contains ECC private key
type ECCPrivateStruct struct {
	Key RetrieveECCPrivateKey `json:"ECCPrivateKey"`
}

//RetrieveECCPrivateKey used to unmarshal ECC private key
type RetrieveECCPrivateKey struct {
	MyPublic ecdsa.PublicKey `json:"PublicKey"`
	MyD      *big.Int        `json:"D"`
}

// Decode read RetrieveECCPrivateKey from json
func (r *RetrieveECCPrivateKey) Decode(data []byte) bool {
	if e := json.Unmarshal(data, r); e != nil {
		return false
	}
	return true
}

// DigitalWalletIp ip and port for digital wallet
type DigitalWalletIp struct {
	DigitalwalletIp   string
	Digitalwalletport string
}

//TempValidator contain validator struct and status of validator
type TempValidator struct {
	ValidatorObjec   ValidatorStruct
	ConfirmationCode string
	CurrentTime      time.Time
}

// DigitalWalletIpObj object of digital wallet
var DigitalWalletIpObj = DigitalWalletIp{}

//NewValidatorObj new object of validator
var NewValidatorObj = ValidatorStruct{}

//CurrentValidator contain all data about current node
var CurrentValidator = ValidatorStruct{}

//ValidatorsLstObj contain all validators
var ValidatorsLstObj []ValidatorStruct

//TempValidatorlst contain validators until it got activated by admin
var TempValidatorlst []TempValidator

//ValidatorAdmin is the first admin that can validate every validator
var ValidatorAdmin admin.AdminStruct
