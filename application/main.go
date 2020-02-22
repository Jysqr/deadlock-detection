package main

import (
	"deadlockdetection/node"
	"fmt"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
	"strconv"
	"time"
)

var (
	numNode       = 8
	thisNode, err = noise.NewNode(noise.WithNodeBindPort(39999))
	network       = kademlia.New()
)

func main() {
	thisNode.Bind(network.Protocol())
	if err := thisNode.Listen(); err != nil {
		panic(err)
	}
	address := thisNode.Addr()
	//var run = true
	//get int from gui
	for i := 0; i < numNode; i++ {
		fmt.Println(strconv.Itoa(i) + " init")
		go func(i int) {
			node.Start(i, address)
		}(i)
	}
	time.Sleep(3 * time.Second)
	fmt.Printf("Boss discovered %d peer(s).\n", len(network.Discover()))
}
