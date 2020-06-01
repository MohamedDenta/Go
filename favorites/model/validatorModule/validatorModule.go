package validatorModule

import (
	"../broadcastTcp"
	"../ledger"

	"../validator"
)

//GetValidatorPublicKey return current node public key
func GetValidatorPublicKey(myvalidator validator.ValidatorStruct, minerIP string) {
	myvalidator.ECCPublicKey = validator.CurrentValidator.ECCPublicKey
	myvalidator.ECCPrivateKey = validator.CurrentValidator.ECCPrivateKey
	myvalidator.EncECCPublic = validator.CurrentValidator.EncECCPublic
	myvalidator.EncECCPriv = validator.CurrentValidator.EncECCPriv
	broadcastTcp.SendObject(myvalidator, myvalidator.ECCPublicKey, minerIP, "Attached public key", minerIP)
}

//ActiveValidator sends ledger to new validator to activate it
func ActiveValidator(myvalidator *validator.ValidatorStruct) {
	(myvalidator).AddValidator()
	(myvalidator).RemoveFromTemp()

	ledObj := ledger.GetLedger()

	broadcastTcp.BoardcastingTCP(*myvalidator, "", "confirmedvalidator")                                                                          // broadcast the validator
	broadcastTcp.SendObject(ledObj, validator.CurrentValidator.ECCPublicKey, "Empty method", "ledger for new node", myvalidator.ValidatorSoketIP) // will POST the ledger here
}
