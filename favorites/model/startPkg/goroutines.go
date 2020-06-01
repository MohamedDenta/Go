package startPkg

import (
	"../compare"

	"../ledger"

	"../heartbeat"

	"../BackUp"
	"../account"

	// "../broadcastHandle"

	// "../broadcastTcp"
	proof "../proofofstake"
	"../systemupdate"
)

func StartGoRoutine() {
	go account.ClearDeadUser()

	go proof.Mining()

	go heartbeat.SendHeartBeat()

	go systemupdate.Update()

	// go broadcastHandle.OpenSocket(Conf.Server.PrivIP + ":" + Conf.Server.SoketPort)
	go BackUp.CreatBackup()

	go ledger.CompareBlockchain()
	go compare.Routine()
	//DeleteRequestRegister  go routine func to delete requested register temp after 20 minute 20 *60 =1200 second
	// go account.DeleteRequestRegister()
	//manage send object time
	// go broadcastTcp.ManageObject()
}
