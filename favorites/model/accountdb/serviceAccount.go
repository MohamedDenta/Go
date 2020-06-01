package accountdb

import (
	"encoding/json"
)

//-----------------------------------------------------------------------------------------------
//---------------------------------------

func GetNamesandPKsForServiceAccount() (values []AccountNameStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AccountStruct
		var newdata2 AccountNameStruct
		json.Unmarshal(value, &newdata)
		if newdata.AccountRole == "service" {
			newdata2.AccountName = newdata.AccountName
			newdata2.AccountIndex = newdata.AccountPublicKey

			values = append(values, newdata2)
		}
	}
	closedatabase()

	return values
}
