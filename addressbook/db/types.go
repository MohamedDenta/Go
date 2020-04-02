package db

type Person struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Contactinfo `json:"contactinfo"`
}
type Contactinfo struct {
	Address string   `json:"address"`
	Phone   []string `json:"phone"`
}

type SearchRequest struct {
	Key string `json:"key"`
}
