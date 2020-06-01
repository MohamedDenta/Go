package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func CreateFile() {

	err := ioutil.WriteFile("test.sh", []byte{45, 45, 97, 64, 71, 66}, 0700)
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}

}

func main() {
	CreateFile()
	b, _ := ioutil.ReadFile("test.sh")
	fmt.Println(string(b))
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
