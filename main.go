package main

import (
	"deadlockdetection/DeadlockNode"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"time"
)

var (
	numNode       = 15
	thisNode, err = noise.NewNode(noise.WithNodeBindPort(39999))
	network       = kademlia.New()
)

func main() {
	thisNode.Bind(network.Protocol())
	if err := thisNode.Listen(); err != nil {
		panic(err)
	}
	initStruct := DeadlockNode.InitStruct{Address: thisNode.Addr()}
	thisNode.Handle(func(ctx noise.HandlerContext) error {
		fmt.Printf("Got a message from  '%s %s'\n", ctx.ID().Address, string(ctx.Data()))
		return nil
	})
	//var run = true
	//get int from gui
	for i := 0; i < numNode; i++ {
		go func(i int) {
			DeadlockNode.NewDeadlockNode(&initStruct)

		}(i)
	}
	time.Sleep(3 * time.Second)
	fmt.Printf("Boss discovered %d peer(s).\n", len(network.Discover()))

}
