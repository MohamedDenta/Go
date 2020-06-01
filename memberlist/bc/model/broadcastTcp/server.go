package broadcastTcp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"../globalPkg"

	"errors"

	"github.com/hashicorp/memberlist"

	//"context"
	//"os"
	"net"
)

func StartMemberlist() {
	// starting up the memberlist
	msgCh := make(chan []byte)
	hostname, _ := os.Hostname()
	d := new(MyDelegate)
	d.msgCh = msgCh

	globalPkg.Conf = memberlist.DefaultLocalConfig()
	globalPkg.Conf.Name = hostname
	globalPkg.Conf.BindPort = 7947 // avoid port confliction
	globalPkg.Conf.AdvertisePort = globalPkg.Conf.BindPort
	// conf.BindAddr      = "3.13.172.219"
	globalPkg.Conf.TCPTimeout = 10 * time.Second
	globalPkg.Conf.Delegate = d
	globalPkg.Conf.ProbeTimeout = 10 * time.Second

	var err error
	globalPkg.ListNodes, err = memberlist.Create(globalPkg.Conf)
	if err != nil {
		log.Fatal(err)
	}

	local := globalPkg.ListNodes.LocalNode()
	globalPkg.ListNodes.Join([]string{
		fmt.Sprintf("%s:%d", local.Addr.To4().String(), local.Port),
	})

	for {
		select {
		case data := <-d.msgCh:
			msg, ok := ParseMyMessage(data)
			if ok != true {
				continue
			}

			fmt.Println(" ----- ----- ----- ")
			if msg.Key == "clientToServerData" {
				// TODO: what to do with the client data? eg. msg.Value
				log.Printf("received msg: key=%s", msg.Key)
			}
		}
	}
}

// getNodeByIP
func getNodeByIP(ipAddr net.IP) *memberlist.Node {

	members := globalPkg.ListNodes.Members()
	for _, node := range members {
		if node.Name == globalPkg.Conf.Name {
			continue // skip self
		}
		if node.Addr.To4().Equal(ipAddr.To4()) {
			return node
		}
	}
	return nil
}

// func retryJoin
func retryJoin(ipAddr net.IP, port uint) error {
	ipWithPort := ipAddr.To4().String() + ":" + string(port)

	var retryCount uint8
	var err error
	for retryCount <= 2 {
		if _, err = globalPkg.ListNodes.Join([]string{ipWithPort}); err != nil {
			retryCount++
			continue
		} else {
			return nil
		}
	}
	return err
}

// getNode
func getNode(ipAddr net.IP, port uint) (*memberlist.Node, error) {
	// get node by ip from memberlist, if not found retry  3 times to join it.
	node := getNodeByIP(ipAddr)
	if node == nil {
		if err := retryJoin(ipAddr, port); err != nil {
			return nil, err
		} else {
			node = getNodeByIP(ipAddr)
		}
	}
	return node, nil
}

// SendToClient
func sendToClient(m *MyMessage, ip string, port uint) error {

	// parsing and checking the ip.
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return errors.New("invalid IP address " + ip)
	}

	node, err := getNode(ipAddr, port)
	if err != nil {
		return err
	}
	if err := globalPkg.ListNodes.SendReliable(node, m.Bytes()); err != nil {
		return err
	}
	fmt.Println("aaa")
	return nil
}

// SendClientID
func SendClientID(ID string, ip string, port uint) error {
	idBytes, err := json.Marshal(ID)
	if err != nil {
		return err
	}

	localNode := globalPkg.ListNodes.LocalNode()
	// make message
	m := new(MyMessage)
	m.FromAddr = localNode.Addr
	m.FromPort = localNode.Port
	m.Key = "serverToClientID"
	m.Value = idBytes

	if err := sendToClient(m, ip, port); err != nil {
		return err
	}
	return nil
}

// SendClientData
func SendClientData(data []byte, ip string, port uint) error {

	localNode := globalPkg.ListNodes.LocalNode()
	// make message
	m := new(MyMessage)
	m.FromAddr = localNode.Addr
	m.FromPort = localNode.Port
	m.Key = "serverToClientData"
	m.Value = data

	if err := sendToClient(m, ip, port); err != nil {
		return err
	}
	return nil
}
