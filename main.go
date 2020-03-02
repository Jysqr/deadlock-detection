package main

import (
	"deadlockdetection/Barrier"
	"deadlockdetection/DeadlockNode"
	"deadlockdetection/MessageTypes"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"sync"
	"time"
)

var (
	numNode       = 3
	nodeSync      = Barrier.NewBarrier(0)
	bossNode, err = noise.NewNode(noise.WithNodeBindPort(39999))
	network       = kademlia.New()
)

func main() {
	bossNode.Bind(network.Protocol())
	if err := bossNode.Listen(); err != nil {
		panic(err)
	}
	initStruct := DeadlockNode.InitStruct{Address: bossNode.Addr()} //contains the init parameters

	bossNode.Handle(func(ctx noise.HandlerContext) error {
		msgObj, _ := ctx.DecodeMessage()
		msg, ok := msgObj.(*MessageTypes.NodeToBoss)
		if !ok {
			return nil
		}
		go messageHandler(msg)
		return nil
	})

	//get int from gui
	nodeSync = Barrier.NewBarrier(numNode)
	for i := 0; i < numNode; i++ { //builds the nodes and gives them their init parameters
		m := nodeSync.Mutex()
		go func(i int, mutex *sync.Mutex) {
			dn := DeadlockNode.NewDeadlockNode(&initStruct, mutex)
			go dn.Start()
		}(i, m)
	}
	bossNode.RegisterMessage(&MessageTypes.BossToNode{}, MessageTypes.UnmarshalBossToNode)
	bossNode.RegisterMessage(&MessageTypes.NodeToBoss{}, MessageTypes.UnmarshalNodeToBoss)
	time.Sleep(3 * time.Second)
	fmt.Printf("Boss discovered %d peer(s).\n", len(network.Discover()))
	_ = nodeSync.Step()
	time.Sleep(2 * time.Second)
	for !nodeSync.Ready {

	}
	_ = nodeSync.Step()
	for !nodeSync.Ready {

	}

}
func messageHandler(msg *MessageTypes.NodeToBoss) {
	fmt.Println(msg.Report)
}
