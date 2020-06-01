package main

import (
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"

	"strconv"
	"time"

	"./model/accountdb"
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

		globalPkg.CookieObject2 = append(globalPkg.CookieObject2, strconv.Itoa(int(time.Now().Unix())))
		globalPkg.CookieObject2 = append(globalPkg.CookieObject2, strconv.Itoa(int(time.Now().Unix())))
		fmt.Println(globalPkg.CookieObject2)

		startPkg.Init()

		startPkg.HandleRequest()
	} else {
		fmt.Println("please enter config file name")
	}
}
