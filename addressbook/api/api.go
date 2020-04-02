package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"../db"
)

//AddContact add contact
func AddContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := new(db.Person)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(data)
	if err != nil {

		w.WriteHeader(http.StatusServiceUnavailable)
		sendJSON, _ := json.Marshal("Invalid request!")
		w.Write(sendJSON)
		return
	}
	id := db.GetLastID()
	if id != "" {
		i, _ := strconv.Atoi(id)
		i = i + 1
		data.ID = strconv.Itoa(i)
	} else {
		data.ID = "1"
	}
	fine := db.Add(*data)
	if !fine {

		w.WriteHeader(http.StatusServiceUnavailable)
		sendJSON, _ := json.Marshal("error in saving data")
		w.Write(sendJSON)
		return
	}
	w.WriteHeader(200)
	sendJSON, _ := json.Marshal("saved successfully")
	w.Write(sendJSON)

}

//Search search by name or phone
func Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := new(db.SearchRequest)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(data)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		sendJSON, _ := json.Marshal("Invalid request!")
		w.Write(sendJSON)
		return
	}
	var p db.Person
	re := regexp.MustCompile(`^(?:(?:\(?(?:00|\+)([1-4]\d\d|[1-9]\d?)\)?)?[\-\.\ \\\/]?)?((?:\(?\d{1,}\)?[\-\.\ \\\/]?){0,})(?:[\-\.\ \\\/]?(?:#|ext\.?|extension|x)[\-\.\ \\\/]?(\d+))?$`)
	if re.MatchString(data.Key) {
		p = db.FindByPhone(data.Key)
	} else {
		p = db.FindByName(data.Key)
	}
	w.WriteHeader(200)
	sendJSON, _ := json.Marshal(p)
	w.Write(sendJSON)
}

// Sort users alphabitically
func Sort(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users := db.GetAll()
	sort.SliceStable(users, func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	w.WriteHeader(200)
	sendJSON, _ := json.Marshal(users)
	w.Write(sendJSON)
}

//GetAllContacts get all contacts
func GetAllContacts(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	sendJSON, _ := json.Marshal(db.GetAll())
	w.Write(sendJSON)
}
