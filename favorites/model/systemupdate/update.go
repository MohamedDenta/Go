package systemupdate

import (
	"fmt"
	"github.com/BurntSushi/toml"

	"time"
	//"strconv"

	"../heartbeat"
)

type UpdateData struct {
	Updatestruct Updatestruct
}
type Updatestruct struct {
	Currentversion float32
	Updateversion  float32
	Updateurl      string
}

func delaySecond(n time.Duration) {
	time.Sleep(n * time.Second)
}



func Update() {

	for {

		delaySecond(300)
		fmt.Println("cheking update !!")


		fmt.Println(">>>>")
		var UpdateDataObj UpdateData
		toml.DecodeFile("./config.toml", &UpdateDataObj)
		fmt.Println(">>>>", UpdateDataObj.Updatestruct.Currentversion)
		fmt.Println(">>>>", UpdateDataObj.Updatestruct.Updateversion)



		if UpdateDataObj.Updatestruct.Updateversion > UpdateDataObj.Updatestruct.Currentversion {
				heartbeat.SendUpdateHeartBeat(UpdateDataObj.Updatestruct.Updateversion, UpdateDataObj.Updatestruct.Updateurl)
				// changes in config file where current version = update version
				fmt.Println("new update found")
			

	} 



		



	}
}
