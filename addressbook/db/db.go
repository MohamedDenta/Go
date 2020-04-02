package db

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

var DB *leveldb.DB
var Open bool

func opendatabase() bool {

	if !Open {
		Open = true
		DBpath := "Database/addressdb"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {

			return false
		}
		return true
	}
	return true

}
func Add(data Person) bool {
	if !Open {
		opendatabase()
	}
	d, _ := json.Marshal(data)
	err := DB.Put([]byte(data.ID), d, nil)
	if err != nil {

		return false
	}
	return true
}
func GetAll() (values []Person) {
	if !Open {
		opendatabase()
	}
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata Person
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	return values
}

func FindByName(id string) (value Person) {
	if !Open {
		opendatabase()
	}
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		b := iter.Value()
		json.Unmarshal(b, &value)
		if value.Name == id {
			break
		}
	}
	return value
}
func FindByPhone(id string) (value Person) {
	if !Open {
		opendatabase()
	}
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		b := iter.Value()
		json.Unmarshal(b, &value)
		if iscontain(value.Contactinfo.Phone, id) {
			break
		}
	}
	return value
}

func iscontain(arr []string, key string) bool {
	for _, v := range arr {
		if v == key {
			return true
		}
	}
	return false
}
func GetLastID() string {
	if !Open {
		opendatabase()
	}
	var value Person
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		if iter.Last() {
			data := iter.Value()
			json.Unmarshal(data, &value)
			break
		}
	}

	return value.ID
}
