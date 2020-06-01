package main

import (
	"time"

	"github.com/kernullist/gowinsvc"
)

type MySvc struct {
	service *gowinsvc.ServiceObject
}

func (mysvc MySvc) Serve(serviceExit <-chan bool) {
	for {
		select {
		case <-serviceExit:
			mysvc.service.OutputDebugString("[MYSERVICE] My Service Exit~~~\n")
			return
		case <-time.After(1 * time.Second):
			mysvc.service.OutputDebugString(
				"[MYSERVICE] Now : %d:%d:%d\n",
				time.Now().Hour(),
				time.Now().Minute(),
				time.Now().Second())
		}
	}
}

func main() {
	mysvc := new(MySvc)
	mysvc.service = gowinsvc.NewService("myservice")
	mysvc.service.StartServe(mysvc)
}
