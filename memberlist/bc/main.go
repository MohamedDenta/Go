package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"

	"strconv"
	"time"

	"./model/accountdb"
	// block "./model/block"
	"./model/globalPkg"
	"./model/heartbeat"
	"./model/startPkg"
)

func ReadConfig() {
	confPath := flag.String("config", "config.toml", "please enter config file name")
	flag.Parse()
	toml.DecodeFile(*confPath, &startPkg.Conf)
}
func main() {
	fmt.Println("version 20191114013")
	ReadConfig()
	if startPkg.Conf.Server.Ip != "" {
		heartbeat.Opendatabase()
		accountdb.Opendatabase()
		globalPkg.EncryptAccount = "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQCzyib/YAf5GcbYLId4ATYc7Vi8crdobaRDL9ztM4CUcm/EAugo\nO2SMk3lSDxIL1sB0mWKs3kWZa+smUH98KuIfpX7vc0PSauQsFb+t3tUyB5ywE2tl\nNYl1TkLwJbdpGO2gdj0qR+aHDTdP9AU6y9DBlMZPyT07KshTqNwNEnynqwIDAQAB\nAoGAanJb2INw9QlP85mZs3F0KnhUO27oLEoOIUFrWn1NuZZvmevmrDtN8vU1tWH6\n20uQsvhFtff72TROC2dJs6hoLDpwZEthl1iaciQO67yja74afhY5A7hgm0qN7+Wm\n6EgBN7swrYETXAFGN8oYGb9kvOEsYZ0RsHhlHZtfOmfVW+ECQQDYg50EltM4vyI9\nI60wTZqQswuIUGUcC/UBc1zU6YxSFvQq/to4fsWj27KENnD0RxBt0EF2Fal5itep\nehGRDG4DAkEA1JP8TwwyBcHl5irDSZQqBtTERepxQ/1kJaOvbJMhi0dcmv4S7a+V\nQZlqDXb2VuhrI5nsAft5SzsYlO58gg9jOQJBAIDnStJyoWqFkPLpjLDXYxCHKHSN\nuMTL8aBdeIViTqKI+/GlLXK5Nx3pLQ0+BF3K+WMHvBF7sBympuNFw7OhvNUCQFbJ\nZATRscpv8vAZHUl41/+Z9delc0CSvsQvI3tsRhGavM/6Urf/Kyxw+b8thjzM/pC2\nUogspsR0CAElrGdc6OECQHOrwCM2Ck9jDCq4K3D2Aph48LTD4dGa1jRVtk5GT86y\nC1LIZUyYIt4VkOY1XpMxAGr5MMH3idIy/q4oPApKn0g=\n-----END RSA PRIVATE KEY-----\n"

		// b := block.FindBlockByKey("000000000000000000000000000000")
		// b.BlockTransactions[0].TransactionOutPut[0].RecieverPublicKey = "1HXGY3nRYhUSQBTpjbqcaijF6bGoBTE4ej"
		// block.AddBlock(b, true)

		globalPkg.CookieObject2 = append(globalPkg.CookieObject2, strconv.Itoa(int(time.Now().Unix())))
		globalPkg.CookieObject2 = append(globalPkg.CookieObject2, strconv.Itoa(int(time.Now().Unix())))
		fmt.Println(globalPkg.CookieObject2)

		startPkg.Init()

		startPkg.HandleRequest()
	} else {
		fmt.Println("please enter config file name")
	}
}
