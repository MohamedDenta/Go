package validator

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"time"

	"../globalPkg"

	"../admin"
	errorpk "../errorpk"
)

/*----------------Validator structure----------------- */
// type ValidatorStruct struct {
// 	ValidatorIP            string
// 	ValidatorSoketIP       string
// 	ValidatorPublicKey     string
// 	ValidatorPrivateKey    string
// 	ValidatorStakeCoins    float64
// 	ValidatorRegisterTime  time.Time
// 	ValidatorActive        bool
// 	ValidatorLastHeartBeat time.Time
// }

/*----------------Validator lst----------------- */
// type ValidatorsLst struct {
// 	ValidatorsLst []Validator
// }
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

var DigitalWalletIpObj = DigitalWalletIp{}
var NewValidatorObj = ValidatorStruct{}
var CurrentValidator = ValidatorStruct{}
var ValidatorsLstObj []ValidatorStruct

//TempValidatorlst contain validators until it got activated by admin
var TempValidatorlst []TempValidator

//ValidatorAdmin is the first admin that can validate every validator
var ValidatorAdmin admin.AdminStruct

//AddValidator function to add validator in the validators list
func AddValidator(validatorObj ValidatorStruct) string {
	validatorObj.Index = NewIndex()
	if validationAdd(validatorObj) {
		ValidatorCreate(validatorObj)
		ValidatorsLstObj = append(ValidatorsLstObj, validatorObj)
		return ""
	}
	return errorpk.AddError("Add Validator validator package", "enter correct validator object not exist before", "hack error")
}

/*----------------function to update validator on the validators list----------------- */
func UpdateValidator(validatorObj ValidatorStruct) string {

	for index, validatorExistsObj := range ValidatorsLstObj {
		if validatorExistsObj.ValidatorPublicKey == validatorObj.ValidatorPublicKey {
			validatorObj.ValidatorPrivateKey = validatorExistsObj.ValidatorPrivateKey
			ValidatorCreate(validatorObj)
			ValidatorsLstObj[index] = validatorObj
			return ""

		}
	}
	return errorpk.AddError("Update Validator validator package", "Can't find the validator object "+validatorObj.ValidatorPublicKey, "hack error")

}

/*----------------function to delete validator from the validators list ----------------- */
func DeleteValidator(validatorObj ValidatorStruct) string {
	for index, validatorExistsObj := range ValidatorsLstObj {
		if validatorExistsObj.ValidatorPublicKey == validatorObj.ValidatorPublicKey {
			ValidatorsLstObj = append(ValidatorsLstObj[:index], ValidatorsLstObj[index+1:]...)
			return ""
		}
	}

	errorpk.AddError("Delete Validator validator package", "Can't find the validator object "+validatorObj.ValidatorPublicKey, "hack error")
	return "Can't find the validator object" + validatorObj.ValidatorPublicKey

}

/*----------------function to validate the validator object ----------------- */
// func (validatorObj *Validator) validation() { //url.Values
// 	// https://medium.com/@thedevsaddam/an-easy-way-to-validate-go-request-c15182fd11b1

// }

func validationAdd(validatorObj ValidatorStruct) bool {
	existAdd := true

	for _, validatorExistsObj := range ValidatorsLstObj {
		if validatorExistsObj.ValidatorIP == validatorObj.ValidatorIP || validatorExistsObj.ValidatorSoketIP == validatorObj.ValidatorSoketIP || validatorExistsObj.ValidatorPublicKey == validatorObj.ValidatorPublicKey {
			errorpk.AddError("validation Add validator package", "The validator object already exists"+validatorExistsObj.ValidatorPublicKey, "hack error")
			existAdd = false
			break
		}
	}
	// if existAdd == true {
	// 	// body := "plz confirm to add Validator "
	// 	// globalPkg.SendEmail(body)
	// }

	return existAdd
}

//FindValidatorByValidatorIP find validator by validator ip
func FindValidatorByValidatorIP(validatorip string) ValidatorStruct {
	validatorObj, _ := findValidatorByIP(validatorip)
	return validatorObj
}

//AddValidatorTemporary add th validator to temp list and send confirmation email to admin
func AddValidatorTemporary(validator TempValidator) {
	TempValidatorlst = append(TempValidatorlst, validator)
	adminslst := admin.GetAllAdmins()
	ValidatorAdmin = adminslst[0]
	SendConfMail(ValidatorAdmin, validator)
}

//SendConfMail send confirmation email to admin
func SendConfMail(admn admin.AdminStruct, validator TempValidator) {

	body := "Dear " + `,
Thank you for joining Inovatian’s InoChain, your request has been processed and your wallet has been created successfully.
Your confirmation code is: ` + `
Please follow the following link to activate your wallet:
(If this link is not clickable, please copy and paste into a new browser)  
` +
		globalPkg.GlobalObj.Downloadfileip + "/ConfirmedValidatorAPI?confirmationcode=" + validator.ConfirmationCode +
		`
 This is a no-reply email; for any enquiries please contact info@inovatian.com
If you did not create this wallet, please disregard this email.
Regards,
Inovatian Team`
	fmt.Println("---------*    Confirmation Code     *--------   ", validator.ConfirmationCode)
	sendEmail(body, admn.AdminEmail)
}

func sendEmail(Body string, Email string) {
	//mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n";
	from := "noreply@inovatian.com" ///// "inovatian.tech@gmail.com"
	pass := "ino13579$"             /////your passward   ////

	to := Email //Email of User

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Inovatian Validator Verification\n\n" + Body

	///confirmation link

	err := smtp.SendMail("mail.inovatian.com:26",
		smtp.PlainAuth("", from, pass, "mail.inovatian.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Println("sent, visit", Email)
}

var randomTable = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

//EncodeToString return confirmation code
func EncodeToString(max int) string {
	buffer := make([]byte, max)
	_, err := io.ReadAtLeast(rand.Reader, buffer, max)
	if err != nil {
		errorpk.AddError("account encodeToString", "the string is more than the max", "runtime error")
	}

code:
	for index := 0; index < len(buffer); index++ {
		buffer[index] = randomTable[int(buffer[index])%len(randomTable)]
	}
	for _, validObjec := range TempValidatorlst {
		if validObjec.ConfirmationCode == string(buffer) {
			goto code
		}
	}
	return string(buffer)
}
