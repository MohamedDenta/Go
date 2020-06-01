package account

import (

	"fmt"
	mathrand "math/rand"

	"strings"
	"time"


	"../accountdb"
	"../globalPkg"
	"../validator"
)
var userobjlst []User

//ClearDeadUser clear User After 5 minite from Registration
func ClearDeadUser() {
	for {
		time.Sleep(time.Second * time.Duration(mathrand.Int31n(globalPkg.GlobalObj.DeleteAccountLoopTimeInseacond)))
		t := globalPkg.UTCtime()
		for index, userObj := range userobjlst {
			t2 := userObj.CurrentTime
			Subtime := (t.Sub(t2)).Seconds()
			if Subtime > globalPkg.GlobalObj.DeleteAccountTimeInseacond { ///globalPkg.GlobalObj.DeleteAccountTimeInseacond {
				fmt.Println(userObj.CurrentTime)
				fmt.Println("subbbbbb", Subtime)
				RemoveUserFromtemp(index)
			}
		}

		for index, userObj := range resetPassReq {
			t2 := userObj.CurrentTime
			if t.Sub(t2).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond { ////globalPkg.GlobalObj.DeleteAccountTimeInseacond
				RemoveResetpassFromtemp(index)
			}
		}
	}
}

//AddUserIntemp to add User in userobj list
func AddUserIntemp(userobj User) {
	fmt.Println(userobj)
	userobjlst = append(userobjlst, userobj)
	fmt.Println("resetPassReq", userobjlst)
}

//AddResetpassObjInTemp Add Update Pass REq to Array
func AddResetpassObjInTemp(ResetpassObj ResetPasswordData) {
	resetPassReq = append(resetPassReq, ResetpassObj)
}

//RemoveUserFromtemp func
func RemoveUserFromtemp(index int) {
	userobjlst = append(userobjlst[:index], userobjlst[index+1:]...)
}

//RemoveResetpassFromtemp func
func RemoveResetpassFromtemp(index int) {
	resetPassReq = append(resetPassReq[:index], resetPassReq[index+1:]...)
}

//SetUserObjLst func
func SetUserObjLst(userObjLst []User) {
	userobjlst = userObjLst
}

//GetUserObjLst func
func GetUserObjLst() []User {
	return userobjlst
}
//NewIndex new index
func NewIndex() string {
	LastIndex := getLastIndex()

	index := 0
	if LastIndex != "-1" {
		// TODO : split LastIndex
		res := strings.Split(LastIndex, "_")
		index = globalPkg.ConvertFixedLengthStringtoInt(res[len(res)-1]) + 1
	}
	timpIndex, _ := globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.StringFixedLength)
	currentIndex := accountdb.GetHash([]byte(validator.CurrentValidator.ValidatorIP)) + "_" + timpIndex
	return currentIndex
}