package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/memberlist"
)

func main() {
	config := memberlist.DefaultLocalConfig()

	list, err := memberlist.Create(config)

	if err != nil {
		panic(err)
	}
	node := list.LocalNode()

	// Create an array of nodes we can join. If you're using a loopback
	// environment you'll need to make sure each node is using its own
	// port. This can be set with the configuration's BindPort field.
	var nodes []string
	nodes = append(nodes, "0.0.0.0:7946")

	if _, err = list.Join(nodes); err != nil {
		panic(err)
	}

	// You can provide a byte representation of any metadata here. You can broadcast the
	// config for each node in some serialized format like JSON. By default, this is
	// limited to 512 bytes, so may not be suitable for large amounts of data.
	node.Meta = []byte("some metadata")
	fmt.Println(node.Addr, " : ", node.Port)

	// Create a channel to listen for exit signals
	stop := make(chan os.Signal, 1)

	// Register the signals we want to be notified, these 3 indicate exit
	// signals, similar to CTRL+C
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	<-stop

	// Leave the cluster with a 5 second timeout. If leaving takes more than 5
	// seconds we return.
	if err := ml.Leave(time.Second * 5); err != nil {
		panic(err)
	}

}
