package errorpk

import (
	"strconv"
	"strings"
	"time"
)

//ConvertTimeToString function to convert time to string
func ConvertTimeToString(timeObj time.Time) string {
	return timeObj.Format("2006-01-02 15:04:05:06 PM -0000")
}

//AddError function to add error in json file
func AddError(functionName string, notes string, typeErr string) string {
	errorObj := ErrorStruct{}
	errorObj.ErrorFunctionName = functionName
	errorObj.ErrorNotes = notes
	errorObj.ErrorType = typeErr

	currentTime, _ := time.Parse("2006-01-02 15:04:05:06 PM -0000", time.Now().UTC().Format("2006-01-02 15:04:05:06 PM -0000"))
	key := errorObj.ErrorFunctionName + "_" + ConvertTimeToString(currentTime) //current time in string
	funcName_Time := key

	//make sure the current time is not equal to the last error time
	//if it does we will add index to it
	now := ConvertTimeToString(currentTime)
	for index := 1; findErrorByKey(key); index++ {
		key = funcName_Time + "_" + strconv.Itoa(index)                    // key = functionName_time_index
		now = ConvertTimeToString(currentTime) + "_" + strconv.Itoa(index) // now = time_index
	}

	errorObj.ErrorTime = now
	errorCreate(errorObj)

	return errorObj.ErrorNotes + " \n"

}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func isEqual(start, end, check time.Time) bool {
	return check.Equal(start) || check.Equal(end)
}

//DeleteErrorsBetweenTowTimes delets all errors happend between specific 2 times
func DeleteErrorsBetweenTowTimes(From string, To string) bool {
	fromTime, _ := time.Parse("2006-01-02 15:04:05:06 PM -0000", From) //convert to time fromate
	toTime, _ := time.Parse("2006-01-02 15:04:05:06 PM -0000", To)     //convert to time fromate

	allErrors := GetAllErrors() //get all errors
	for _, err := range allErrors {
		t := strings.Split(err.ErrorTime, "_")
		errTime, _ := time.Parse("2006-01-02 15:04:05:06 PM -0000", t[0])                //get the time in time formate without _ if exist
		if inTimeSpan(fromTime, toTime, errTime) || isEqual(fromTime, toTime, errTime) { //if time in range or equal
			errKey := err.ErrorFunctionName + "_" + err.ErrorTime //get the error key
			del := ErrorDelete(errKey)                            //delete the error
			if del == false {
				return false
			}
		}
	}
	return true
}
